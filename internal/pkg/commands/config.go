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
	"fmt"

	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"gopkg.in/yaml.v2"
)

const copyrightNotice = `
# Copyright 2019-present Ciena Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
`

const (
	DEFAULT_CONFIG_LIST_FORMAT = "table{{if .Current}}*{{end}}\t{{.Name}}"
)

type ConfigShow struct {
	All bool `short:"a" long:"all" description:"show all configuration for all stacks"`
}

type ConfigStackListOutput struct {
	Current bool
	Name    string
}

type ConfigStackList struct {
	ListOutputOptions
}

type ConfigStackUse struct {
	Args struct {
		Name string `positional-arg-name:"STACK_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type ConfigStackAdd struct {
	Args struct {
		Name string `positional-arg-name:"STACK_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type ConfigStackRemove struct {
	Args struct {
		Name string `positional-arg-name:"STACK_NAME" required:"yes"`
	} `positional-args:"yes"`
}

type ConfigOpts struct {
	Show  ConfigShow `command:"show"`
	Stack struct {
		List   ConfigStackList   `command:"list"`
		Use    ConfigStackUse    `command:"use"`
		Add    ConfigStackAdd    `command:"add"`
		Remove ConfigStackRemove `command:"remove"`
	} `command:"stack"`
}

func RegisterConfigCommands(parent *flags.Parser) {
	if command, err := parent.AddCommand("config", "configuration opperations", "Commands to display and modify voltctl configuration", &ConfigOpts{}); err != nil {
		Error.Fatalf("Unexpected error while attempting to register config commands : %s", err)
	} else {
		command.SubcommandsOptional = true
	}
}

func (options *ConfigShow) Execute(args []string) error {
	//GlobalConfig
	ProcessConfigWithOptions(ProcessConfigOptions{
		HonorStackArgument: true,
	})
	var b []byte
	var err error
	if options.All {
		b, err = yaml.Marshal(GlobalConfig)
	} else {
		if GlobalOptions.Stack != "" {
			b, err = yaml.Marshal(GlobalConfig.GetStack(GlobalOptions.Stack))
		} else {
			b, err = yaml.Marshal(GlobalConfig.GetCurrentStack())
		}
	}
	if err != nil {
		return err
	}
	fmt.Println(copyrightNotice)
	fmt.Println(string(b))
	return nil
}

func (options *ConfigStackAdd) Execute(args []string) error {
	ProcessConfigWithOptions(ProcessConfigOptions{
		AddStack:         true,
		StackName:        options.Args.Name,
		ProcessOverrides: true})
	return WriteGlobalOptions()
}

func (options *ConfigStackRemove) Execute(args []string) error {
	ProcessConfigWithOptions(ProcessConfigOptions{
		ProcessOverrides: false})
	for i, stack := range GlobalConfig.Stacks {
		if stack.Name == options.Args.Name {
			GlobalConfig.Stacks = append(GlobalConfig.Stacks[:i], GlobalConfig.Stacks[i+1:]...)
			if GlobalConfig.CurrentStack == stack.Name {
				GlobalConfig.CurrentStack = ""
			}
			return WriteGlobalOptions()
		}
	}
	return nil
}

func (options *ConfigStackUse) Execute(args []string) error {
	ProcessConfig()
	for _, stack := range GlobalConfig.Stacks {
		if stack.Name == options.Args.Name {
			GlobalConfig.CurrentStack = stack.Name
			return WriteGlobalOptions()
		}
	}
	return nil
}

func (options *ConfigStackList) Execute(args []string) error {
	ProcessConfigWithOptions(ProcessConfigOptions{
		ProcessOverrides: false})
	var out []ConfigStackListOutput
	for _, stack := range GlobalConfig.Stacks {
		out = append(out, ConfigStackListOutput{
			Name:    stack.Name,
			Current: GlobalConfig.CurrentStack == stack.Name,
		})
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("config-list", "format", DEFAULT_CONFIG_LIST_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Name}}"
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("config-list", "order", "")
	}

	// Make sure json output prints an empty list, not "null"
	if len(out) == 0 {
		out = make([]ConfigStackListOutput, 0)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      out,
	}

	GenerateOutput(&result)
	return nil

}
