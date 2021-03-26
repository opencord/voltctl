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
	"os"
	"path"
	"testing"

	flags "github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
)

// Test that ProcessGlobalOptions does not interfere with GlobalConfig
// default.
func TestProcessGlobalOptionsWithDefaults(t *testing.T) {
	os.Setenv("VOLTCONFIG", "__DOES_NOT_EXIST__")

	parser := flags.NewNamedParser(path.Base(os.Args[0]), flags.Default|flags.PassAfterNonOption)
	_, err := parser.AddGroup("Global Options", "", &GlobalOptions)
	assert.Nil(t, err, "unexpected error adding group")
	RegisterConfigCommands(parser)
	_, err = parser.ParseArgs([]string{"config"})
	assert.Nil(t, err, "unexpected error paring arguments")
	ProcessGlobalOptions()

	assert.Equal(t, "localhost:55555", GlobalConfig.Current().Server, "wrong default hostname for server")
}

func TestParseSize(t *testing.T) {
	var res uint64
	var err error

	res, err = parseSize("8M")
	assert.Nil(t, err)
	assert.Equal(t, uint64(8388608), res)

	res, err = parseSize("8MB")
	assert.Nil(t, err)
	assert.Equal(t, uint64(8388608), res)

	res, err = parseSize("8MiB")
	assert.Nil(t, err)
	assert.Equal(t, uint64(8388608), res)

	res, err = parseSize("foobar")
	assert.NotNil(t, err)
	assert.Equal(t, uint64(0), res)
}
