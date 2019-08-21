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
package main

import (
	"github.com/ciena/voltctl/internal/pkg/commands"
	flags "github.com/jessevdk/go-flags"
	"os"
	"path"
)

func main() {

	parser := flags.NewNamedParser(path.Base(os.Args[0]), flags.Default|flags.PassAfterNonOption)
	_, err := parser.AddGroup("Global Options", "", &commands.GlobalOptions)
	if err != nil {
		panic(err)
	}
	commands.RegisterAdapterCommands(parser)
	commands.RegisterDeviceCommands(parser)
	commands.RegisterLogicalDeviceCommands(parser)
	commands.RegisterDeviceGroupCommands(parser)
	commands.RegisterVersionCommands(parser)
	commands.RegisterCompletionCommands(parser)
	commands.RegisterConfigCommands(parser)
	commands.RegisterComponentCommands(parser)

	_, err = parser.ParseArgs(os.Args[1:])
	if err != nil {
		_, ok := err.(*flags.Error)
		if ok {
			real := err.(*flags.Error)
			if real.Type == flags.ErrHelp {
				return
			}
		} else {
			panic(err)
		}
		os.Exit(1)
	}
}
