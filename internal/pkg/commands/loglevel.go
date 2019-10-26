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
	"fmt"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/model"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

type SetLogLevelOutput struct {
	ComponentName string
	Status        string
	Error         string
}

type SetLogLevelOpts struct {
	OutputOptions
	Package string `short:"p" long:"package" description:"Package name to set log level"`
	Args    struct {
		Level     string
		Component []string
	} `positional-args:"yes" required:"yes"`
}

type GetLogLevelsOpts struct {
	ListOutputOptions
	Args struct {
		Component []string
	} `positional-args:"yes" required:"yes"`
}

type ListLogLevelsOpts struct {
	ListOutputOptions
}

type LogLevelOpts struct {
	SetLogLevel   SetLogLevelOpts   `command:"set"`
	GetLogLevels  GetLogLevelsOpts  `command:"get"`
	ListLogLevels ListLogLevelsOpts `command:"list"`
}

var logLevelOpts = LogLevelOpts{}

const (
	DEFAULT_LOGLEVELS_FORMAT   = "table{{ .ComponentName }}\t{{.PackageName}}\t{{.Level}}"
	DEFAULT_SETLOGLEVEL_FORMAT = "table{{ .ComponentName }}\t{{.Status}}\t{{.Error}}"
)

func RegisterLogLevelCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("loglevel", "loglevel commands", "Get and set log levels", &logLevelOpts)
	if err != nil {
		Error.Fatalf("Unable to register log level commands with voltctl command parser: %s", err.Error())
	}
}

func MapListAppend(m map[string][]string, name string, item string) {
	list, okay := m[name]
	if okay {
		m[name] = append(list, item)
	} else {
		m[name] = []string{item}
	}
}

/*
 * A roundabout way of going of using the LogLevel enum to map from
 * a string to an integer that we can pass into the dynamic
 * proto.
 *
 * TODO: There's probably an easier way.
 */
func LogLevelStringToInt(logLevelString string) (int32, error) {
	ProcessGlobalOptions() // required for GetMethod()

	/*
	 * Use GetMethod() to get us a descriptor on the proto file we're
	 * interested in.
	 */

	descriptor, _, err := GetMethod("update-log-level")
	if err != nil {
		return 0, err
	}

	/*
	 * Map string LogLevel to enumerated type LogLevel
	 * We have descriptor from above, which is a DescriptorSource
	 * We can use FindSymbol to get at the message
	 */

	loggingSymbol, err := descriptor.FindSymbol("common.LogLevel")
	if err != nil {
		return 0, err
	}

	/*
	 * LoggingSymbol is a Descriptor, but not a MessageDescrptior,
	 * so we can't look at it's fields yet. Go back to the file,
	 * call FindMessage to get the Message, then we can get the
	 * embedded enum.
	 */

	loggingFile := loggingSymbol.GetFile()
	logLevelMessage := loggingFile.FindMessage("common.LogLevel")
	logLevelEnumType := logLevelMessage.GetNestedEnumTypes()[0]
	enumLogLevel := logLevelEnumType.FindValueByName(logLevelString)

	if enumLogLevel == nil {
		return 0, fmt.Errorf("Unknown log level %s", logLevelString)
	}

	return enumLogLevel.GetNumber(), nil
}

// Validate a list of component names and throw an error if any one of them is bad.
func ValidateComponentNames(kube_to_arouter map[string][]string, names []string) error {
	var badNames []string
	for _, name := range names {
		_, ok := kube_to_arouter[name]
		if !ok {
			badNames = append(badNames, name)
		}
	}

	if len(badNames) > 0 {
		allowedNames := make([]string, len(kube_to_arouter))
		i := 0
		for k := range kube_to_arouter {
			allowedNames[i] = k
			i++
		}

		return fmt.Errorf("Unknown component(s): %s.\n  (Allowed values for component names: \n    %s)",
			strings.Join(badNames, ", "),
			strings.Join(allowedNames, ",\n    "))
	} else {
		return nil
	}
}

func BuildKubernetesNameMap() (map[string][]string, map[string]string, error) {
	kube_to_arouter := make(map[string][]string)
	arouter_to_kube := make(map[string]string)

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", GlobalOptions.K8sConfig)
	if err != nil {
		return nil, nil, err
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/part-of=voltha",
	})
	if err != nil {
		return nil, nil, err
	}

	if len(pods.Items) == 0 {
		return nil, nil, fmt.Errorf("No Voltha pods found in Kubernetes -- verify pod is setup")
	}

	for _, pod := range pods.Items {
		app, ok := pod.Labels["app"]
		if !ok {
			continue
		}

		var arouter_name string

		switch app {
		case "voltha-api-server":
			/*
			 * Assumes a single api_server for now.
			 * TODO: Make labeling changes in charts to be able to derive name from labels
			 */
			arouter_name = "api_server0.api_server01"
		case "rw-core":
			affinity_group, ok := pod.Labels["affinity-group"]
			if !ok {
				Warn.Printf("rwcore %s lacks affinity-group label", pod.Name)
				continue
			}
			affinity_group_core_id, ok := pod.Labels["affinity-group-core-id"]
			if !ok {
				Warn.Printf("rwcore %s lacks affinity-group-core-id label", pod.Name)
				continue
			}
			arouter_name = "vcore" + affinity_group + ".vcore" + affinity_group + affinity_group_core_id
		case "ro-core":
			/*
			 * Assumes a single rocore for now.
			 * TODO: Make labeling changes in charts to be able to derive name from labels
			 */
			arouter_name = "ro_vcore0.ro_vcore01"
		default:
			// skip this pod as it's not relevant
			continue
		}

		// Multiple ways to identify the component

		// 1) The pod name. One pod name maps to exactly one pod.

		arouter_to_kube[arouter_name] = pod.Name
		MapListAppend(kube_to_arouter, pod.Name, arouter_name)

		// 2) The kubernetes component name. A single component (i.e. "core") may map to multiple pods.

		component, ok := pod.Labels["app.kubernetes.io/component"]
		if ok {
			MapListAppend(kube_to_arouter, component, arouter_name)
		}

		// 3) The voltha app label. A single app (i.e. "rwcore") may map to multiple pods.

		MapListAppend(kube_to_arouter, app, arouter_name)

	}

	return kube_to_arouter, arouter_to_kube, nil
}

func (options *SetLogLevelOpts) Execute(args []string) error {
	if len(options.Args.Component) == 0 {
		return fmt.Errorf("Please specify at least one component")
	}

	kube_to_arouter, arouter_to_kube, err := BuildKubernetesNameMap()
	if err != nil {
		return err
	}

	var output []SetLogLevelOutput

	// Validate component names, throw error now to avoid doing partial work
	err = ValidateComponentNames(kube_to_arouter, options.Args.Component)
	if err != nil {
		return err
	}

	// Validate and map the logLevel string to an integer, throw error now to avoid doing partial work
	intLogLevel, err := LogLevelStringToInt(options.Args.Level)
	if err != nil {
		return err
	}

	for _, kubeComponentName := range options.Args.Component {
		var descriptor grpcurl.DescriptorSource
		var conn *grpc.ClientConn
		var method string

		componentNameList := kube_to_arouter[kubeComponentName]

		for _, componentName := range componentNameList {
			conn, err = NewConnection()
			if err != nil {
				return err
			}
			defer conn.Close()
			if strings.HasPrefix(componentName, "api_server") {
				// apiserver's UpdateLogLevel is in the afrouter.Configuration gRPC package
				descriptor, method, err = GetMethod("apiserver-update-log-level")
			} else {
				descriptor, method, err = GetMethod("update-log-level")
			}
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
			defer cancel()

			ll := make(map[string]interface{})
			ll["component_name"] = componentName
			ll["package_name"] = options.Package
			ll["level"] = intLogLevel

			h := &RpcEventHandler{
				Fields: map[string]map[string]interface{}{"common.Logging": ll},
			}
			err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
			if err != nil {
				return err
			}

			if h.Status != nil && h.Status.Err() != nil {
				output = append(output, SetLogLevelOutput{ComponentName: arouter_to_kube[componentName], Status: "Failure", Error: h.Status.Err().Error()})
				continue
			}

			output = append(output, SetLogLevelOutput{ComponentName: arouter_to_kube[componentName], Status: "Success"})
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("loglevel-set", "format", DEFAULT_SETLOGLEVEL_FORMAT)
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

func (options *GetLogLevelsOpts) getLogLevels(methodName string, args []string) error {
	if len(options.Args.Component) == 0 {
		return fmt.Errorf("Please specify at least one component")
	}

	kube_to_arouter, arouter_to_kube, err := BuildKubernetesNameMap()
	if err != nil {
		return err
	}

	var data []model.LogLevel

	// Validate component names, throw error now to avoid doing partial work
	err = ValidateComponentNames(kube_to_arouter, options.Args.Component)
	if err != nil {
		return err
	}

	for _, kubeComponentName := range options.Args.Component {
		var descriptor grpcurl.DescriptorSource
		var conn *grpc.ClientConn
		var method string

		componentNameList := kube_to_arouter[kubeComponentName]

		for _, componentName := range componentNameList {
			conn, err = NewConnection()
			if err != nil {
				return err
			}
			defer conn.Close()
			if strings.HasPrefix(componentName, "api_server") {
				// apiserver's UpdateLogLevel is in the afrouter.Configuration gRPC package
				descriptor, method, err = GetMethod("apiserver-get-log-levels")
			} else {
				descriptor, method, err = GetMethod("get-log-levels")
			}
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
			defer cancel()

			ll := make(map[string]interface{})
			ll["component_name"] = componentName

			h := &RpcEventHandler{
				Fields: map[string]map[string]interface{}{"common.LoggingComponent": ll},
			}
			err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
			if err != nil {
				return err
			}

			if h.Status != nil && h.Status.Err() != nil {
				return h.Status.Err()
			}

			d, err := dynamic.AsDynamicMessage(h.Response)
			if err != nil {
				return err
			}
			items, err := d.TryGetFieldByName("items")
			if err != nil {
				return err
			}

			for _, item := range items.([]interface{}) {
				logLevel := model.LogLevel{}
				logLevel.PopulateFrom(item.(*dynamic.Message))
				logLevel.ComponentName = arouter_to_kube[logLevel.ComponentName]

				data = append(data, logLevel)
			}
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault(methodName, "format", DEFAULT_LOGLEVELS_FORMAT)
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault(methodName, "order", "")
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

func (options *GetLogLevelsOpts) Execute(args []string) error {
	return options.getLogLevels("loglevel-get", args)
}

func (options *ListLogLevelsOpts) Execute(args []string) error {
	var getOptions GetLogLevelsOpts
	var podNames []string

	_, arouter_to_kube, err := BuildKubernetesNameMap()
	if err != nil {
		return err
	}

	for _, podName := range arouter_to_kube {
		podNames = append(podNames, podName)
	}

	// Just call GetLogLevels with a list of podnames that includes everything relevant.

	getOptions.ListOutputOptions = options.ListOutputOptions
	getOptions.Args.Component = podNames

	return getOptions.getLogLevels("loglevel-list", args)
}
