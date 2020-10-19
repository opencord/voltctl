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

	assert.Equal(t, "localhost:55555", GlobalConfig.Server, "wrong default hostname for server")
}

func TestSplitHostPort(t *testing.T) {
	data := []struct {
		name        string
		endpoint    string
		defaultHost string
		defaultPort int
		host        string
		port        int
		err         bool
	}{
		{"Host and port specified", "host:1234", "default", 4321, "host", 1234, false},
		{"Host only specified", "host", "default", 4321, "host", 4321, false},
		{"Host: only specified", "host:", "default", 4321, "host", 4321, false},
		{"Port only specified", ":1234", "default", 4321, "default", 1234, false},
		{"Colon only", ":", "default", 4321, "default", 4321, false},
		{"Empty endpoint", "", "default", 4321, "default", 4321, false},
		{"IPv4 and port specified", "1.2.3.4:1234", "4.3.2.1", 4321, "1.2.3.4", 1234, false},
		{"IPv4 only specified", "1.2.3.4", "4.3.2.1", 4321, "1.2.3.4", 4321, false},
		{"IPv4: only specified", "1.2.3.4:", "4.3.2.1", 4321, "1.2.3.4", 4321, false},
		{"IPv4 Port only specified", ":1234", "4.3.2.1", 4321, "4.3.2.1", 1234, false},
		{"IPv4 Colon only", ":", "4.3.2.1", 4321, "4.3.2.1", 4321, false},
		{"IPv4 Empty endpoint", "", "4.3.2.1", 4321, "4.3.2.1", 4321, false},
		{"IPv6 and port specified", "[0001:c0ff:eec0::::ffff]:1234", "0001:c0ff:eec0::::aaaa", 4321, "0001:c0ff:eec0::::ffff", 1234, false},
		{"IPv6 only specified", "[0001:c0ff:eec0::::ffff]", "0001:c0ff:eec0::::aaaa", 4321, "0001:c0ff:eec0::::ffff", 4321, false},
		{"IPv6: only specified", "[0001:c0ff:eec0::::ffff]:", "0001:c0ff:eec0::::aaaa", 4321, "0001:c0ff:eec0::::ffff", 4321, false},
		{"IPv6 Port only specified", ":1234", "0001:c0ff:eec0::::aaaa", 4321, "0001:c0ff:eec0::::aaaa", 1234, false},
		{"IPv6 Colon only", ":", "0001:c0ff:eec0::::aaaa", 4321, "0001:c0ff:eec0::::aaaa", 4321, false},
		{"IPv6 Empty endpoint", "", "0001:c0ff:eec0::::aaaa", 4321, "0001:c0ff:eec0::::aaaa", 4321, false},
		{"Invalid port", "host:1b", "default", 4321, "", 0, true},
		{"Too many colons", "ho:st:1b", "default", 4321, "", 0, true},
		{"IPv4 Invalid port", "1.2.3.4:1b", "4.3.2.1", 4321, "", 0, true},
		{"IPv4 Too many colons", "1.2.3.4::1234", "4.3.2.1", 4321, "", 0, true},
		{"IPv6 Invalid port", "[0001:c0ff:eec0::::ffff]:1b", "0001:c0ff:eec0::::aaaa", 4321, "", 0, true},
		{"IPv6 Too many colons", "0001:c0ff:eec0::::ffff:1234", "0001:c0ff:eec0::::aaaa", 4321, "", 0, true},
	}

	for _, args := range data {
		t.Run(args.name, func(t *testing.T) {
			h, p, err := splitEndpoint(args.endpoint, args.defaultHost, args.defaultPort)
			if args.err {
				assert.NotNil(t, err, "unexpected non-error result")
			} else {
				assert.Nil(t, err, "unexpected error result")
			}
			if !args.err && err == nil {
				assert.Equal(t, args.host, h, "unexpected host value")
				assert.Equal(t, args.port, p, "unexpected port value")
			}
		})
	}
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

/*func TestDeviceUpdateList_Execute(t *testing.T) {
	ul := DeviceUpdateList{}
	ul.Args.Id = "9fae2b0b-e7b6-451a-83f2-13b14129a2f7"
	ul.Filter = "Timestamp>2020-11-12T16:49:19Z"
	ul.Execute([]string{})
}*/
