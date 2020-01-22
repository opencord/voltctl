/*
 * Copyright 2019-present Open Networking Foundation
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
	flags "github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestDefaultVersion(t *testing.T) {
	os.Setenv("VOLTCONFIG", "__DOES_NOT_EXIST__")

	parser := flags.NewNamedParser(path.Base(os.Args[0]), flags.Default|flags.PassAfterNonOption)
	_, err := parser.AddGroup("Global Options", "", &GlobalOptions)
	assert.Nil(t, err, "unexpected error adding group")
	RegisterConfigCommands(parser)
	_, err = parser.ParseArgs([]string{"config"})
	assert.Nil(t, err, "unexpected error paring arguments")
	ProcessGlobalOptions()

	assert.Equal(t, "v3", GlobalConfig.ApiVersion, "wrong default version for API version")
}
