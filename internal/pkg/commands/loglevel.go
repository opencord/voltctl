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
	"errors"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/model"
	"github.com/opencord/voltha-lib-go/v3/pkg/config"
	"github.com/opencord/voltha-lib-go/v3/pkg/db/kvstore"
	"github.com/opencord/voltha-lib-go/v3/pkg/log"
	"strconv"
	"strings"
)

const (
	defaultKVStoreType    = "etcd"
	defaultKVStoreTimeout = 5 //in seconds
	defaultKVStoreHost    = "127.0.0.1"
	defaultKVStorePort    = 2379 // Consul = 8500; Etcd = 2379
	defaultComponentName  = "global"
	defaultPackageName    = "default"
)

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
		Level     string
		Component []string
	} `positional-args:"yes" required:"yes"`
}

// ListLogLevelOpts represents the supported CLI arguments for the loglevel list command
type ListLogLevelsOpts struct {
	ListOutputOptions
	Args struct {
		Component []string
	} `positional-args:"yes" required:"yes"`
}

// ClearLogLevelOpts represents the supported CLI arguments for the loglevel clear command
type ClearLogLevelsOpts struct {
	OutputOptions
	Args struct {
		Component []string
	} `positional-args:"yes" required:"yes"`
}

// LogLevelOpts represents the loglevel commands
type LogLevelOpts struct {
	SetLogLevel    SetLogLevelOpts    `command:"set"`
	ListLogLevels  ListLogLevelsOpts  `command:"list"`
	ClearLogLevels ClearLogLevelsOpts `command:"clear"`
}

var logLevelOpts = LogLevelOpts{}

const (
	DEFAULT_LOGLEVELS_FORMAT       = "table{{ .ComponentName }}\t{{.PackageName}}\t{{.Level}}"
	DEFAULT_LOGLEVEL_RESULT_FORMAT = "table{{ .ComponentName }}\t{{.Status}}\t{{.Error}}"
)

// RegisterLogLevelCommands is used to  register set,list and clear loglevel of components
func RegisterLogLevelCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("loglevel", "loglevel commands", "list,set and clear log levels of components", &logLevelOpts)
	if err != nil {
		Error.Fatalf("Unable to register log level commands with voltctl command parser: %s", err.Error())
	}
}

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
		if strings.Contains(component, "#") {
			val := strings.Split(component, "#")
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

	if options.Args.Level != "" {
		if _, err := log.StringToLogLevel(options.Args.Level); err != nil {
			return fmt.Errorf("Unknown log level %s. Allowed values are <INFO>,<DEBUG>,<ERROR>,<WARN>,<FATAL>", options.Args.Level)
		}
	}

	logLevelConfig, err = processComponentListArgs(options.Args.Component)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	client, err := kvstore.NewEtcdClient(defaultKVStoreHost+":"+strconv.Itoa(defaultKVStorePort), defaultKVStoreTimeout)
	defer client.Close()
	if err != nil {
		return fmt.Errorf("Unable to create kvstore client %s", err)
	}
	cm := config.NewConfigManager(client, defaultKVStoreType, defaultKVStoreHost, defaultKVStorePort, defaultKVStoreTimeout)

	var output []LogLevelOutput

	for _, lConfig := range logLevelConfig {

		logConfig := cm.InitComponentConfig(lConfig.ComponentName, config.ConfigTypeLogLevel)
		err := logConfig.Save(context.Background(), lConfig.PackageName, strings.ToUpper(options.Args.Level))
		if err != nil {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Failure", Error: err.Error()})
		} else {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Success"})
		}

	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("loglevel-set", "format", DEFAULT_LOGLEVEL_RESULT_FORMAT)
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
		data           []model.LogLevel
		componentList  []string
		logLevelConfig map[string]string
		err            error
	)

	client, err := kvstore.NewEtcdClient(defaultKVStoreHost+":"+strconv.Itoa(defaultKVStorePort), defaultKVStoreTimeout)
	defer client.Close()
	if err != nil {
		return fmt.Errorf("Unable to create kvstore client %s", err)
	}
	cm := config.NewConfigManager(client, defaultKVStoreType, defaultKVStoreHost, defaultKVStorePort, defaultKVStoreTimeout)

	if len(options.Args.Component) == 0 {
		componentList, err = cm.RetrieveComponentList(context.Background(), config.ConfigTypeLogLevel)
		if err != nil {
			return fmt.Errorf("Unable to retrieve list of voltha components : %s ", err)
		}
	} else {
		for _, component := range options.Args.Component {
			componentList = append(componentList, component)
		}
	}

	for _, componentName := range componentList {
		logConfig := cm.InitComponentConfig(componentName, config.ConfigTypeLogLevel)

		logLevelConfig, err = logConfig.RetrieveAll(context.Background())
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
		outputFormat = GetCommandOptionWithDefault("loglevel-list", "format", DEFAULT_LOGLEVELS_FORMAT)
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("loglevel-list", "order", "")
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

	logLevelConfig, err = processComponentListArgs(options.Args.Component)
	if err != nil {
		return fmt.Errorf("%s", err)
	}

	client, err := kvstore.NewEtcdClient(defaultKVStoreHost+":"+strconv.Itoa(defaultKVStorePort), defaultKVStoreTimeout)
	defer client.Close()
	if err != nil {
		return fmt.Errorf("Unable to create kvstore client %s", err)
	}
	cm := config.NewConfigManager(client, defaultKVStoreType, defaultKVStoreHost, defaultKVStorePort, defaultKVStoreTimeout)

	var output []LogLevelOutput
	for _, lConfig := range logLevelConfig {

		logConfig := cm.InitComponentConfig(lConfig.ComponentName, config.ConfigTypeLogLevel)

		err := logConfig.Delete(context.Background(), lConfig.PackageName)
		if err != nil {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Failure", Error: err.Error()})
		} else {
			output = append(output, LogLevelOutput{ComponentName: lConfig.ComponentName, Status: "Success"})
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("loglevel-clear", "format", DEFAULT_LOGLEVEL_RESULT_FORMAT)
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
