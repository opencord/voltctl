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
	"os"

	flags "github.com/jessevdk/go-flags"
	configv3 "github.com/opencord/voltctl/internal/pkg/apis/config/v3"
	"github.com/opencord/voltctl/pkg/format"
	yaml "gopkg.in/yaml.v2"
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

type StackInfo struct {
	Current string `json:"current"`
	Name    string `json:"name"`
	Server  string `json:"server"`
	Kafka   string `json:"kafka"`
	KvStore string `json:"kvstore"`
}

func write() error {
	w, err := os.OpenFile(GlobalOptions.Config, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer w.Close()
	encode := yaml.NewEncoder(w)
	if err := encode.Encode(GlobalConfig); err != nil {
		return err
	}
	return nil
}

func (options *StackList) Execute(args []string) error {

	ReadConfig()
	ApplyOptionOverrides(nil)

	var data []StackInfo
	for _, stack := range GlobalConfig.Stacks {
		s := StackInfo{
			Name:    stack.Name,
			Server:  stack.Server,
			Kafka:   stack.Kafka,
			KvStore: stack.KvStore,
		}
		if stack.Name == GlobalConfig.CurrentStack {
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

	ReadConfig()

	for _, stack := range GlobalConfig.Stacks {
		if stack.Name == options.Args.Name {
			GlobalConfig.CurrentStack = stack.Name
			ApplyOptionOverrides(stack)
			if err := write(); err != nil {
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

	ReadConfig()
	ApplyOptionOverrides(nil)

	for i, stack := range GlobalConfig.Stacks {
		if stack.Name == options.Args.Name {
			GlobalConfig.Stacks = append(GlobalConfig.Stacks[:i], GlobalConfig.Stacks[i+1:]...)
			if GlobalConfig.CurrentStack == stack.Name {
				GlobalConfig.CurrentStack = ""
			}
			if err := write(); err != nil {
				Error.Fatal(err.Error())
			}
			fmt.Printf("wrote: '%s'\n", GlobalOptions.Config)
			return nil
		}
	}

	Error.Fatalf("stack not found, '%s'", options.Args.Name)
	return nil
}

func (options *StackAdd) Execute(args []string) error {

	ReadConfig()
	stack := GlobalConfig.StackByName(options.Args.Name)

	if stack == nil {
		stack = configv3.NewDefaultStack(options.Args.Name)
		GlobalConfig.Stacks = append(GlobalConfig.Stacks, stack)
	}
	if GlobalConfig.CurrentStack == "" {
		GlobalConfig.CurrentStack = options.Args.Name
	}
	ApplyOptionOverrides(stack)
	if err := write(); err != nil {
		Error.Fatal(err.Error())
	}
	return nil
}
