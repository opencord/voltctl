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
	"github.com/opencord/voltctl/internal/pkg/completion"
)

type BashOptions struct{}

type CompletionOptions struct {
	BashOptions `command:"bash"`
}

func RegisterCompletionCommands(parent *flags.Parser) {
	if _, err := parent.AddCommand("completion", "generate shell compleition", "Commands to generate shell compleition information", &CompletionOptions{}); err != nil {
		Error.Fatalf("Unexpected error while attempting to register completion commands : %s", err)
	}
}

func (options *BashOptions) Execute(args []string) error {
	fmt.Print(completion.Bash)
	return nil
}
