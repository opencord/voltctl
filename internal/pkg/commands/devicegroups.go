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
	"github.com/golang/protobuf/ptypes/empty"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltha-protos/v4/go/voltha"
)

const (
	DEFAULT_DEVICE_GROUP_FORMAT = "table{{ .Id }}\t{{.LogicalDevices}}\t{{.Devices}}"
)

type DeviceGroupList struct {
	ListOutputOptions
}

type DeviceGroupOpts struct {
	List DeviceGroupList `command:"list"`
}

var deviceGroupOpts = DeviceGroupOpts{}

func RegisterDeviceGroupCommands(parser *flags.Parser) {
	if _, err := parser.AddCommand("devicegroup", "device group commands", "Commands to query and manipulate VOLTHA device groups",
		&deviceGroupOpts); err != nil {
		Error.Fatalf("Unexpected error while attempting to register device group commands : %s", err)
	}
}

func (options *DeviceGroupList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	deviceGroups, err := client.ListDeviceGroups(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("device-group-list", "format", DEFAULT_DEVICE_GROUP_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("device-group-list", "order", "")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      deviceGroups.Items,
	}

	GenerateOutput(&result)
	return nil
}
