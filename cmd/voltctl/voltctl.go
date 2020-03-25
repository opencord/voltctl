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
	"fmt"
	"os"
	"path"

	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/internal/pkg/commands"
)

func main() {

	/*
	 * When completion is invoked the environment variable GO_FLAG_COMPLETION
	 * is set. The go-flags library uses this as a key to do completion when
	 * parsing the arguments. The problem is that when doing compleition the
	 * library doesn't bother setting the arguments into the structures. As
	 * such voltctl's configuration options, in partitular VOLTCONFIG, is
	 * not set and thus connection to voltha servers can not be established
	 * and completion for device ID etc will not work.
	 *
	 * An issue against go-flags has been opened, but as a work around if or
	 * until it is fixed there is a bit of a hack. Unset the environment var
	 * parse the global config options, and then continue with normal
	 * completion.
	 */
	compval := os.Getenv("GO_FLAGS_COMPLETION")
	if len(compval) > 0 {
		os.Unsetenv("GO_FLAGS_COMPLETION")
		pp := flags.NewNamedParser(path.Base(os.Args[0]), flags.HelpFlag|flags.PassDoubleDash|flags.PassAfterNonOption)
		if _, err := pp.AddGroup("Global Options", "", &commands.GlobalOptions); err != nil {
			commands.Error.Fatalf("Unable to set up global options for command completion: %s", err.Error())
		}
		if _, err := pp.Parse(); err != nil {
			commands.Error.Fatalf("Unable to parse command line options for command completion: %s", err.Error())
		}
		os.Setenv("GO_FLAGS_COMPLETION", compval)
	}

	parser := flags.NewNamedParser(path.Base(os.Args[0]), flags.HelpFlag|flags.PassDoubleDash|flags.PassAfterNonOption)
	_, err := parser.AddGroup("Global Options", "", &commands.GlobalOptions)
	if err != nil {
		commands.Error.Fatalf("Unable to parse global command options: %s", err.Error())
	}
	commands.RegisterAdapterCommands(parser)
	commands.RegisterDeviceCommands(parser)
	commands.RegisterLogicalDeviceCommands(parser)
	commands.RegisterDeviceGroupCommands(parser)
	commands.RegisterVersionCommands(parser)
	commands.RegisterCompletionCommands(parser)
	commands.RegisterConfigCommands(parser)
	commands.RegisterComponentCommands(parser)
	commands.RegisterLogLevelCommands(parser)
	commands.RegisterEventCommands(parser)

	_, err = parser.Parse()
	if err != nil {
		_, ok := err.(*flags.Error)
		if ok {
			real := err.(*flags.Error)
			// Send help message to stdout, else to stderr
			if real.Type == flags.ErrHelp {
				parser.WriteHelp(os.Stdout)
			} else {
				if real.Error() != commands.NoReportErr.Error() {
					fmt.Fprintf(os.Stderr, "%s\n", real.Error())
				}
				os.Exit(1)
			}
			return
		} else if err != commands.NoReportErr {
			commands.Error.Fatal(commands.ErrorToString(err))
		}
		os.Exit(1)
	}
}
