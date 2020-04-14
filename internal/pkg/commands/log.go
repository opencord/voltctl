/*
 * Copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/model"
	"github.com/opencord/voltha-lib-go/v3/pkg/config"
	"github.com/opencord/voltha-lib-go/v3/pkg/db/kvstore"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultComponentName = "global"
	defaultPackageName   = "default"
	logPackagesListKey   = "log_package_list" // kvstore key containing list of allowed log packages
)

// Custom Option representing <component-name>#<package-name> format (package is optional)
// This is used by 'log level set' commands
type ComponentAndPackageName string

// Custom Option representing currently configured log configuration in <component-name>#<package-name> format (package is optional)
// This is used by 'log level clear' commands
type ConfiguredComponentAndPackageName string

// Custom Option representing component-name. This is used by 'log level list' and 'log package list' commands
type ComponentName string

// Custom Option representing Log Level (one of debug, info, warn, error, fatal)
type LevelName string

// LogLevelOutput represents the  output structure for the loglevel
type LogLevelOutput struct {
	ComponentName string
	Status        string
	Error         string
}

// SetLogLevelOpts represents the supported CLI arguments for the loglevel set command
type SetLogLevelOpts struct {
	OutputOptions
	Args struct {
		Level     LevelName
		Component []ComponentAndPackageName
	} `positional-args:"yes" required:"yes"`
}

// ListLogLevelOpts represents the supported CLI arguments for the loglevel list command
type ListLogLevelsOpts struct {
	ListOutputOptions
	Args struct {
		Component []ComponentName
	} `positional-args:"yes" required:"yes"`
}

// ClearLogLevelOpts represents the supported CLI arguments for the loglevel clear command
type ClearLogLevelsOpts struct {
	OutputOptions
	Args struct {
		Component []ConfiguredComponentAndPackageName
	} `positional-args:"yes" required:"yes"`
}

// ListLogLevelOpts represents the supported CLI arguments for the loglevel list command
type ListLogPackagesOpts struct {
	ListOutputOptions
	Args struct {
		Component []ComponentName
	} `positional-args:"yes" required:"yes"`
}

// LogPackageOpts represents the log package commands
type LogPackageOpts struct {
	ListLogPackages ListLogPackagesOpts `command:"list"`
}

// LogLevelOpts represents the log level commands
type LogLevelOpts struct {
	SetLogLevel    SetLogLevelOpts    `command:"set"`
	ListLogLevels  ListLogLevelsOpts  `command:"list"`
	ClearLogLevels ClearLogLevelsOpts `command:"clear"`
}

// LogOpts represents the log commands
type LogOpts struct {
	LogLevel   LogLevelOpts   `command:"level"`
	LogPackage LogPackageOpts `command:"package"`
}

var logOpts = LogOpts{}

const (
	DEFAULT_LOG_LEVELS_FORMAT   = "table{{ .ComponentName }}\t{{.PackageName}}\t{{.Level}}"
	DEFAULT_LOG_PACKAGES_FORMAT = "table{{ .ComponentName }}\t{{.PackageName}}"
	DEFAULT_LOG_RESULT_FORMAT   = "table{{ .ComponentName }}\t{{.Status}}\t{{.Error}}"
)

func toStringArray(arg interface{}) []string {
	var list []string
	if cnl, ok := arg.([]ComponentName); ok {
		for _, cn := range cnl {
			list = append(list, string(cn))
		}
	} else if cpnl, ok := arg.([]ComponentAndPackageName); ok {
		for _, cpn := range cpnl {
			list = append(list, string(cpn))
		}
	} else if ccpnl, ok := arg.([]ConfiguredComponentAndPackageName); ok {
		for _, ccpn := range ccpnl {
			list = append(list, string(ccpn))
		}
	}

	return list
}

// RegisterLogCommands is used to register log and its sub-commands e.g. level, package etc
func RegisterLogCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("log", "log config commands", "list, set, clear log levels and list packages of components", &logOpts)
	if err != nil {
		Error.Fatalf("Unable to register log commands with voltctl command parser: %s", err.Error())
	}
}

// Common method to get list of VOLTHA components using k8s API. Used for validation and auto-complete
// Just return a blank list in case of any error
func getVolthaComponentNames() []string {
	var componentList []string

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", GlobalOptions.K8sConfig)
	if err != nil {
		// Ignore any error
		return componentList
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return componentList
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/part-of=voltha",
	})
	if err != nil {
		return componentList
	}

	for _, pod := range pods.Items {
		componentList = append(componentList, pod.ObjectMeta.Labels["app.kubernetes.io/name"])
	}

	return componentList
}

func getPackageNames(componentName string) ([]string, error) {
	list := []string{defaultPackageName}

	ProcessGlobalOptions()

	log.SetAllLogLevel(log.FatalLevel)

	client, err := kvstore.NewEtcdClient(GlobalConfig.KvStore, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()), log.FatalLevel)
	if err != nil {
		return nil, fmt.Errorf("Unable to create kvstore client %s", err)
	}
	defer client.Close()

	// Already error checked during option processing
	host, port, _ := splitEndpoint(GlobalConfig.KvStore, defaultKvHost, defaultKvPort)
	cm := config.NewConfigManager(client, supportedKvStoreType, host, port, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()))

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.KvStoreConfig.Timeout)
	defer cancel()

	componentMetadata := cm.InitComponentConfig(componentName, config.ConfigTypeMetadata)

	value, err := componentMetadata.Retrieve(ctx, logPackagesListKey)
	if err != nil || value == "" {
		return list, nil
	}

	var packageList []string
	if err = json.Unmarshal([]byte(value), &packageList); err != nil {
		return list, nil
	}

	list = append(list, packageList...)

	return list, nil
}

func (ln *LevelName) Complete(match string) []flags.Completion {
	levels := []string{"debug", "info", "warn", "error", "fatal"}

	var list []flags.Completion
	for _, name := range levels {
		if strings.HasPrefix(name, strings.ToLower(match)) {
			list = append(list, flags.Completion{Item: name})
		}
	}

	return list
}

func (cpn *ComponentAndPackageName) Complete(match string) []flags.Completion {

	componentNames := getVolthaComponentNames()

	// Return nil if no component names could be fetched
	if len(componentNames) == 0 {
		return nil
	}

	// Check to see if #was specified, and if so, we know we have
	// to split component name and package
	parts := strings.SplitN(match, "#", 2)

	var list []flags.Completion
	for _, name := range componentNames {
		if strings.HasPrefix(name, parts[0]) {
			list = append(list, flags.Completion{Item: name})
		}
	}

	// If the possible completions > 1 then we have to stop here
	// as we can't suggest packages
	if len(parts) == 1 || len(list) > 1 {
		return list
	}

	// Ok, we have a valid, unambiguous component name and there
	// is a package separator, so lets try to expand the package
	// and in this case we will replace the list we have so
	// far with the new list
	cname := list[0].Item
	base := []flags.Completion{{Item: fmt.Sprintf("%s#%s", cname, parts[1])}}
	packages, err := getPackageNames(cname)
	if err != nil || len(packages) == 0 {
		return base
	}

	list = []flags.Completion{}
	for _, pname := range packages {
		if strings.HasPrefix(pname, parts[1]) {
			list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s", cname, pname)})
		}
	}

	// if package part is present and still no match found based on prefix, user may be using
	// short-hand notation for package name (last element of package path string e.g. kafka).
	// Attempt prefix match against last element of package path (after the last / character)
	if len(list) == 0 && len(parts[1]) >= 3 {
		var mplist []string
		for _, pname := range packages {
			pnameparts := strings.Split(pname, "/")
			if strings.HasPrefix(pnameparts[len(pnameparts)-1], parts[1]) {
				mplist = append(mplist, pname)
			}
		}

		// add to completion list if only a single match is found
		if len(mplist) == 1 {
			list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s", cname, mplist[0])})
		}
	}

	// If the component name was expanded but package name match was not found, list will still be empty
	// We should return entry with just component name auto-completed and package name unchanged.
	if len(list) == 0 && cname != parts[0] {
		// Returning 2 entries with <completed-component-name>#<package-part> as prefix
		// Just 1 entry will auto-complete the argument
		list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s1", cname, parts[1])})
		list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s2", cname, parts[1])})
	}

	return list
}

func (ccpn *ConfiguredComponentAndPackageName) Complete(match string) []flags.Completion {

	var list []flags.Completion

	ProcessGlobalOptions()

	log.SetAllLogLevel(log.FatalLevel)

	client, err := kvstore.NewEtcdClient(GlobalConfig.KvStore, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()), log.FatalLevel)
	if err != nil {
		return list
	}
	defer client.Close()

	// Already error checked during option processing
	host, port, _ := splitEndpoint(GlobalConfig.KvStore, defaultKvHost, defaultKvPort)
	cm := config.NewConfigManager(client, supportedKvStoreType, host, port, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()))

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.KvStoreConfig.Timeout)
	defer cancel()

	var componentNames []string
	componentNames, err = cm.RetrieveComponentList(ctx, config.ConfigTypeLogLevel)

	// Return nil if no component names could be fetched
	if err != nil || len(componentNames) == 0 {
		return nil
	}

	// Check to see if #was specified, and if so, we know we have
	// to split component name and package
	parts := strings.SplitN(match, "#", 2)

	for _, name := range componentNames {
		if strings.HasPrefix(name, parts[0]) {
			list = append(list, flags.Completion{Item: name})

			// Handle scenario when one component is exact substring of other e.g. read-write-cor
			// is substring of read-write-core (last e missing). Such a wrong component name
			// can get configured during log level set operation
			// In case of exact match of component name, use it if package part is present
			if name == parts[0] && len(parts) == 2 {
				list = []flags.Completion{flags.Completion{Item: name}}
				break
			}
		}
	}

	// If the possible completions > 1 then we have to stop here
	// as we can't suggest packages
	if len(parts) == 1 || len(list) > 1 {
		return list
	}

	// Ok, we have a valid, unambiguous component name and there
	// is a package separator, so lets try to expand the package
	// and in this case we will replace the list we have so
	// far with the new list
	cname := list[0].Item
	base := []flags.Completion{{Item: fmt.Sprintf("%s#%s", cname, parts[1])}}

	// Get list of packages configured for matching component name
	logConfig := cm.InitComponentConfig(cname, config.ConfigTypeLogLevel)
	logLevels, err1 := logConfig.RetrieveAll(ctx)
	if err1 != nil || len(logLevels) == 0 {
		return base
	}

	packages := make([]string, len(logLevels))
	list = []flags.Completion{}
	for pname, _ := range logLevels {
		pname = strings.ReplaceAll(pname, "#", "/")
		packages = append(packages, pname)
		if strings.HasPrefix(pname, parts[1]) {
			list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s", cname, pname)})
		}
	}

	// if package part is present and still no match found based on prefix, user may be using
	// short-hand notation for package name (last element of package path string e.g. kafka).
	// Attempt prefix match against last element of package path (after the last / character)
	if len(list) == 0 && len(parts[1]) >= 3 {
		var mplist []string
		for _, pname := range packages {
			pnameparts := strings.Split(pname, "/")
			if strings.HasPrefix(pnameparts[len(pnameparts)-1], parts[1]) {
				mplist = append(mplist, pname)
			}
		}

		// add to completion list if only a single match is found
		if len(mplist) == 1 {
			list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s", cname, mplist[0])})
		}
	}

	// If the component name was expanded but package name match was not found, list will still be empty
	// We should return entry with just component name auto-completed and package name unchanged.
	if len(list) == 0 && cname != parts[0] {
		// Returning 2 entries with <completed-component-name>#<package-part> as prefix
		// Just 1 entry will auto-complete the argument
		list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s1", cname, parts[1])})
		list = append(list, flags.Completion{Item: fmt.Sprintf("%s#%s2", cname, parts[1])})
	}

	return list
}

func (cn *ComponentName) Complete(match string) []flags.Completion {

	componentNames := getVolthaComponentNames()

	// Return nil if no component names could be fetched
	if len(componentNames) == 0 {
		return nil
	}

	var list []flags.Completion
	for _, name := range componentNames {
		if strings.HasPrefix(name, match) {
			list = append(list, flags.Completion{Item: name})
		}
	}

	return list
}

// Return nil if no component names could be fetched
// processComponentListArgs stores  the component name and package names given in command arguments to LogLevel
// It checks the given argument has # key or not, if # is present then split the argument for # then stores first part as component name
// and second part as package name
func processComponentListArgs(Components []string) ([]model.LogLevel, error) {

	var logLevelConfig []model.LogLevel

	if len(Components) == 0 {
		Components = append(Components, defaultComponentName)
	}

	for _, component := range Components {
		logConfig := model.LogLevel{}
		val := strings.SplitN(component, "#", 2)

		if strings.Contains(val[0], "/") {
			return nil, errors.New("the component name '" + val[0] + "' contains an invalid character '/'")
		}

		if len(val) > 1 {
			if val[0] == defaultComponentName {
				return nil, errors.New("global level doesn't support packageName")
			}
			logConfig.ComponentName = val[0]
			logConfig.PackageName = strings.ReplaceAll(val[1], "/", "#")
		} else {
			logConfig.ComponentName = component
			logConfig.PackageName = defaultPackageName
		}
		logLevelConfig = append(logLevelConfig, logConfig)
	}
	return logLevelConfig, nil
}

// This method set loglevel for components.
// For example, using below command loglevel can be set for specific component with default packageName
// voltctl loglevel set level  <componentName>
// For example, using below command loglevel can be set for specific component with specific packageName
// voltctl loglevel set level <componentName#packageName>
// For example, using below command loglevel can be set for more than one component for default package and other component for specific packageName
// voltctl loglevel set level <componentName1#packageName> <componentName2>
func (options *SetLogLevelOpts) Execute(args []string) error {
	var (
		logLevelConfig []model.LogLevel
		err            error
	)
	ProcessGlobalOptions()

	log.SetAllLogLevel(log.FatalLevel)

	if options.Args.Level != "" {
		if _, err := log.StringToLogLevel(string(options.Args.Level)); err != nil {
			return fmt.Errorf("Unknown log level '%s'. Allowed values are  DEBUG, INFO, WARN, ERROR, FATAL", options.Args.Level)
		}
	}

	logLevelConfig, err = processComponentListArgs(toStringArray(options.Args.Component))
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	client, err := kvstore.NewEtcdClient(GlobalConfig.KvStore, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()), log.FatalLevel)
	if err != nil {
		return fmt.Errorf("Unable to create kvstore client %s", err)
	}
	defer client.Close()

	// Already error checked during option processing
	host, port, _ := splitEndpoint(GlobalConfig.KvStore, defaultKvHost, defaultKvPort)
	cm := config.NewConfigManager(client, supportedKvStoreType, host, port, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()))

	var output []LogLevelOutput

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.KvStoreConfig.Timeout)
	defer cancel()

	for _, lConfig := range logLevelConfig {

		logConfig := cm.InitComponentConfig(lConfig.ComponentName, config.ConfigTypeLogLevel)
		err := logConfig.Save(ctx, lConfig.PackageName, strings.ToUpper(string(options.Args.Level)))
		if err != nil {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Failure", Error: err.Error()})
		} else {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Success"})
		}

	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("log-level-set", "format", DEFAULT_LOG_RESULT_FORMAT)
	}
	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      output,
	}

	GenerateOutput(&result)
	return nil
}

// This method list loglevel for components.
// For example, using below command loglevel can be list for specific component
// voltctl loglevel list  <componentName>
// For example, using below command loglevel can be list for all the components with all the packageName
// voltctl loglevel list
func (options *ListLogLevelsOpts) Execute(args []string) error {

	var (
		// Initialize to empty as opposed to nil so that -o json will
		// display empty list and not null VOL-2742
		data           []model.LogLevel = []model.LogLevel{}
		componentList  []string
		logLevelConfig map[string]string
		err            error
	)
	ProcessGlobalOptions()

	log.SetAllLogLevel(log.FatalLevel)

	client, err := kvstore.NewEtcdClient(GlobalConfig.KvStore, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()), log.FatalLevel)
	if err != nil {
		return fmt.Errorf("Unable to create kvstore client %s", err)
	}
	defer client.Close()

	// Already error checked during option processing
	host, port, _ := splitEndpoint(GlobalConfig.KvStore, defaultKvHost, defaultKvPort)
	cm := config.NewConfigManager(client, supportedKvStoreType, host, port, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()))

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.KvStoreConfig.Timeout)
	defer cancel()

	if len(options.Args.Component) == 0 {
		componentList, err = cm.RetrieveComponentList(ctx, config.ConfigTypeLogLevel)
		if err != nil {
			return fmt.Errorf("Unable to retrieve list of voltha components : %s ", err)
		}
	} else {
		componentList = toStringArray(options.Args.Component)
	}

	for _, componentName := range componentList {
		logConfig := cm.InitComponentConfig(componentName, config.ConfigTypeLogLevel)

		logLevelConfig, err = logConfig.RetrieveAll(ctx)
		if err != nil {
			return fmt.Errorf("Unable to retrieve loglevel configuration for component %s : %s", componentName, err)
		}

		for packageName, level := range logLevelConfig {
			logLevel := model.LogLevel{}
			if packageName == "" {
				continue
			}

			pName := strings.ReplaceAll(packageName, "#", "/")
			logLevel.PopulateFrom(componentName, pName, level)
			data = append(data, logLevel)
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("log-level-list", "format", DEFAULT_LOG_LEVELS_FORMAT)
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("log-level-list", "order", "")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}
	GenerateOutput(&result)
	return nil
}

// This method clear loglevel for components.
// For example, using below command loglevel can be clear for specific component with default packageName
// voltctl loglevel clear  <componentName>
// For example, using below command loglevel can be clear for specific component with specific packageName
// voltctl loglevel clear <componentName#packageName>
func (options *ClearLogLevelsOpts) Execute(args []string) error {

	var (
		logLevelConfig []model.LogLevel
		err            error
	)
	ProcessGlobalOptions()

	log.SetAllLogLevel(log.FatalLevel)

	logLevelConfig, err = processComponentListArgs(toStringArray(options.Args.Component))
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	client, err := kvstore.NewEtcdClient(GlobalConfig.KvStore, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()), log.FatalLevel)
	if err != nil {
		return fmt.Errorf("Unable to create kvstore client %s", err)
	}
	defer client.Close()

	// Already error checked during option processing
	host, port, _ := splitEndpoint(GlobalConfig.KvStore, defaultKvHost, defaultKvPort)
	cm := config.NewConfigManager(client, supportedKvStoreType, host, port, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()))

	var output []LogLevelOutput

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.KvStoreConfig.Timeout)
	defer cancel()

	for _, lConfig := range logLevelConfig {

		if lConfig.ComponentName == defaultComponentName {
			return fmt.Errorf("The global default loglevel cannot be cleared.")
		}

		logConfig := cm.InitComponentConfig(lConfig.ComponentName, config.ConfigTypeLogLevel)

		err := logConfig.Delete(ctx, lConfig.PackageName)
		if err != nil {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Failure", Error: err.Error()})
		} else {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Success"})
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("log-level-clear", "format", DEFAULT_LOG_RESULT_FORMAT)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      output,
	}

	GenerateOutput(&result)
	return nil
}

// This method lists registered log packages for components.
// For example, available log packages can be listed for specific component using below command
// voltctl loglevel listpackage  <componentName>
// For example, available log packages can be listed for all the components using below command (omitting component name)
// voltctl loglevel listpackage
func (options *ListLogPackagesOpts) Execute(args []string) error {

	var (
		// Initialize to empty as opposed to nil so that -o json will
		// display empty list and not null VOL-2742
		data          []model.LogLevel = []model.LogLevel{}
		componentList []string
		err           error
	)

	ProcessGlobalOptions()

	log.SetAllLogLevel(log.FatalLevel)

	client, err := kvstore.NewEtcdClient(GlobalConfig.KvStore, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()), log.FatalLevel)
	if err != nil {
		return fmt.Errorf("Unable to create kvstore client %s", err)
	}
	defer client.Close()

	// Already error checked during option processing
	host, port, _ := splitEndpoint(GlobalConfig.KvStore, defaultKvHost, defaultKvPort)
	cm := config.NewConfigManager(client, supportedKvStoreType, host, port, int(GlobalConfig.KvStoreConfig.Timeout.Seconds()))

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.KvStoreConfig.Timeout)
	defer cancel()

	if len(options.Args.Component) == 0 {
		componentList, err = cm.RetrieveComponentList(ctx, config.ConfigTypeLogLevel)
		if err != nil {
			return fmt.Errorf("Unable to retrieve list of voltha components : %s ", err)
		}

		// Include default global package as well when displaying packages for all components
		logLevel := model.LogLevel{}
		logLevel.PopulateFrom(defaultComponentName, defaultPackageName, "")
		data = append(data, logLevel)
	} else {
		for _, name := range options.Args.Component {
			componentList = append(componentList, string(name))
		}
	}

	for _, componentName := range componentList {
		componentMetadata := cm.InitComponentConfig(componentName, config.ConfigTypeMetadata)

		value, err := componentMetadata.Retrieve(ctx, logPackagesListKey)
		if err != nil || value == "" {
			// Ignore any error in retrieval for log package list; some components may not store it
			continue
		}

		var packageList []string
		if err = json.Unmarshal([]byte(value), &packageList); err != nil {
			continue
		}

		for _, packageName := range packageList {
			logLevel := model.LogLevel{}
			logLevel.PopulateFrom(componentName, packageName, "")
			data = append(data, logLevel)
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("log-package-list", "format", DEFAULT_LOG_PACKAGES_FORMAT)
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("log-package-list", "order", "ComponentName,PackageName")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}
	GenerateOutput(&result)
	return nil
}
