/*
 * Copyright 2021-present Ciena Corporation
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
	"github.com/opencord/voltctl/pkg/model"
)

const (
	DefaultOutputFormat = "table{{.Current}}\t{{.Name}}\t{{.Server}}\t{{.KvStore}}\t{{.Kafka}}"
)

type StackUse struct {
	Args struct {
		Name string `positional-arg-name:"NAME" required:"yes"`
	} `positional-args:"yes"`
}

type StackAdd struct {
	Args struct {
		Name string `positional-arg-name:"NAME" required:"yes"`
	} `positional-args:"yes"`
}

type StackDelete struct {
	Args struct {
		Name string `positional-arg-name:"NAME" required:"yes"`
	} `positional-args:"yes"`
}

type StackList struct {
	ListOutputOptions
}

type StackOptions struct {
	List   StackList   `command:"list" description:"list all configured stacks"`
	Add    StackAdd    `command:"add" description:"add or update the named stack using command line options"`
	Delete StackDelete `command:"delete" description:"delete the specified stack configuration"`
	Use    StackUse    `command:"use" description:"perist the specified stack to be used by default"`
}

func RegisterStackCommands(parent *flags.Parser) {
	if _, err := parent.AddCommand("stack", "generate voltctl configuration", "Commands to generate voltctl configuration", &StackOptions{}); err != nil {
		Error.Fatalf("Unexpected error while attempting to register config commands : %s", err)
	}
}

func (options *StackList) Execute(args []string) error {

	ProcessGlobalOptions()

	var data []model.Stack
	for _, stack := range GlobalConfig.GetStacks() {
		s := model.Stack{
			Name:    stack.GetName(),
			Server:  stack.GetServer(),
			Kafka:   stack.GetKafka(),
			KvStore: stack.GetKvStore(),
		}
		if stack.GetName() == GlobalConfig.GetCurrentStack() {
			s.Current = "*"
		}
		data = append(data, s)
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("stack-list", "format",
			DefaultOutputFormat)
	}
	if options.Quiet {
		outputFormat = "{{.Name}}"
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("stack-list", "order", "")
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

func (options *StackUse) Execute(args []string) error {

	ProcessGlobalOptions()
	for _, stack := range GlobalConfig.GetStacks() {
		if stack.GetName() == options.Args.Name {
			if err := GlobalConfig.SetCurrentStack(options.Args.Name); err != nil {
				Error.Fatal(err.Error())
			}
			if err := GlobalConfig.Write(GlobalOptions.Config); err != nil {
				Error.Fatal(err.Error())
			}
			fmt.Printf("wrote: '%s'\n", GlobalOptions.Config)
			return nil
		}
	}

	Error.Fatalf("unknown stack: '%s'", options.Args.Name)

	return nil
}

func (options *StackDelete) Execute(args []string) error {

	ProcessGlobalOptions()
	if err := GlobalConfig.DeleteStack(options.Args.Name); err != nil {
		Error.Fatal(err.Error())
	}
	if err := GlobalConfig.Write(GlobalOptions.Config); err != nil {
		Error.Fatal(err.Error())
	}
	fmt.Printf("wrote: '%s'\n", GlobalOptions.Config)
	return nil
}

func (options *StackAdd) Execute(args []string) error {
	ProcessGlobalOptionsToStack(options.Args.Name)
	if GlobalConfig.GetCurrentStack() == "" {
		_ = GlobalConfig.SetCurrentStack(options.Args.Name)
	}
	if err := GlobalConfig.Write(GlobalOptions.Config); err != nil {
		Error.Fatal(err.Error())
	}
	return nil
}
