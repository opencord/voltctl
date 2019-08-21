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
)

type SetLogLevelOutput struct {
	ComponentName string
	Status        string
	Error         string
}

type SetLogLevelOpts struct {
	OutputOptions
	Package string `short:"p" long:"package" description:"Package name to set filter level"`
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
		panic(err)
	}
}

func ValidateComponentName(s string) error {
	switch s {
	case "api-server":
		return nil
	case "rwcore":
		return nil
	// rocore currently unsupported
	// ofagent currently unsupported
	default:
		return fmt.Errorf("Invalid component name %s", s)
	}
}

func (options *SetLogLevelOpts) Execute(args []string) error {
	if len(options.Args.Component) == 0 {
		return fmt.Errorf("Please specify at least one component")
	}

	var output []SetLogLevelOutput

	for _, componentName := range options.Args.Component {
		var descriptor grpcurl.DescriptorSource
		var conn *grpc.ClientConn
		var method string

		err := ValidateComponentName(componentName)
		if err != nil {
			return err
		}

		if componentName == "api-server" {
			conn, err = NewAffinityConnection()
			if err != nil {
				return err
			}
			defer conn.Close()
			descriptor, method, err = GetMethod("affinity-update-log-level")
			if err != nil {
				return err
			}
		} else {
			conn, err = NewConnection()
			if err != nil {
				return err
			}
			defer conn.Close()
			descriptor, method, err = GetMethod("update-log-level")
			if err != nil {
				return err
			}
		}

		/* Map string LogLevel to enumerated type LogLevel
		We have descriptor from above, which is a DescriptorSource
		We can use FindSymbol to get at the message
		*/

		loggingSymbol, err := descriptor.FindSymbol("common.LogLevel")
		if err != nil {
			return err
		}

		/* LoggingSymbol is a Descriptor, but not a MessageDescrptior,
		so we can't look at it's fields yet. Go back to the file,
		call FindMessage to get the Message, then we can get the
		embedded enum.
		*/

		loggingFile := loggingSymbol.GetFile()
		logLevelMessage := loggingFile.FindMessage("common.LogLevel")
		logLevelEnumType := logLevelMessage.GetNestedEnumTypes()[0]
		enumLogLevel := logLevelEnumType.FindValueByName(options.Args.Level).GetNumber()

		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
		defer cancel()

		ll := make(map[string]interface{})
		ll["component_name"] = componentName
		ll["package_name"] = options.Package
		ll["level"] = enumLogLevel // Options.Args.Level

		h := &RpcEventHandler{
			Fields: map[string]map[string]interface{}{"common.Logging": ll},
		}
		err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
		if err != nil {
			return err
		}

		if h.Status != nil && h.Status.Err() != nil {
			output = append(output, SetLogLevelOutput{ComponentName: componentName, Status: "Failure", Error: h.Status.Err().Error()})
			continue
		}

		output = append(output, SetLogLevelOutput{ComponentName: componentName, Status: "Success"})
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_SETLOGLEVEL_FORMAT
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

func (options *GetLogLevelsOpts) Execute(args []string) error {
	if len(options.Args.Component) == 0 {
		return fmt.Errorf("Please specify at least one component")
	}

	var data []model.LogLevel

	for _, componentName := range options.Args.Component {
		var descriptor grpcurl.DescriptorSource
		var conn *grpc.ClientConn
		var method string

		err := ValidateComponentName(componentName)
		if err != nil {
			return err
		}

		if componentName == "api-server" {
			conn, err = NewAffinityConnection()
			if err != nil {
				return err
			}
			defer conn.Close()
			descriptor, method, err = GetMethod("affinity-get-log-levels")
			if err != nil {
				return err
			}
		} else {
			conn, err = NewConnection()
			if err != nil {
				return err
			}
			defer conn.Close()
			descriptor, method, err = GetMethod("get-log-levels")
			if err != nil {
				return err
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
		defer cancel()

		h := &RpcEventHandler{}
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

			data = append(data, logLevel)
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_LOGLEVELS_FORMAT
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   options.OrderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}
	GenerateOutput(&result)
	return nil
}

func (options *ListLogLevelsOpts) Execute(args []string) error {
	var getOptions GetLogLevelsOpts

	getOptions.ListOutputOptions = options.ListOutputOptions
	getOptions.Args.Component = []string{"api-server"}

	return getOptions.Execute(args)
}
