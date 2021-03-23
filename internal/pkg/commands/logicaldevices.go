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
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltha-protos/v4/go/openflow_13"
	"github.com/opencord/voltha-protos/v4/go/voltha"
	"strings"
)

const (
	DEFAULT_LOGICAL_DEVICE_FORMAT         = "table{{ .Id }}\t{{printf \"%016x\" .DatapathId}}\t{{.RootDeviceId}}\t{{.Desc.SerialNum}}\t{{.SwitchFeatures.NBuffers}}\t{{.SwitchFeatures.NTables}}\t{{printf \"0x%08x\" .SwitchFeatures.Capabilities}}"
	DEFAULT_LOGICAL_DEVICE_PORT_FORMAT    = "table{{.Id}}\t{{.DeviceId}}\t{{.DevicePortNo}}\t{{.RootPort}}\t{{.OfpPortStats.PortNo}}\t{{.OfpPort.HwAddr}}\t{{.OfpPort.Name}}\t{{printf \"0x%08x\" .OfpPort.State}}\t{{printf \"0x%08x\" .OfpPort.Curr}}\t{{.OfpPort.CurrSpeed}}"
	DEFAULT_LOGICAL_DEVICE_INSPECT_FORMAT = `ID: {{.Id}}
  DATAPATHID: {{.DatapathId}}
  ROOTDEVICEID: {{.RootDeviceId}}
  SERIALNUMNER: {{.Desc.SerialNum}}`
)

type LogicalDeviceId string

type LogicalDeviceList struct {
	ListOutputOptions
}

type LogicalDeviceFlowList struct {
	ListOutputOptions
	FlowIdOptions
	Args struct {
		Id LogicalDeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type LogicalDevicePortList struct {
	ListOutputOptions
	Args struct {
		Id LogicalDeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type LogicalDeviceInspect struct {
	OutputOptionsJson
	Args struct {
		Id LogicalDeviceId `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type LogicalDeviceOpts struct {
	List  LogicalDeviceList     `command:"list"`
	Flows LogicalDeviceFlowList `command:"flows"`
	Port  struct {
		List LogicalDevicePortList `command:"list"`
	} `command:"port"`
	Inspect LogicalDeviceInspect `command:"inspect"`
}

var logicalDeviceOpts = LogicalDeviceOpts{}

func RegisterLogicalDeviceCommands(parser *flags.Parser) {
	if _, err := parser.AddCommand("logicaldevice", "logical device commands", "Commands to query and manipulate VOLTHA logical devices", &logicalDeviceOpts); err != nil {
		Error.Fatalf("Unexpected error while attempting to register logical device commands : %s", err)
	}
}

func (i *LogicalDeviceId) Complete(match string) []flags.Completion {
	conn, err := NewConnection()
	if err != nil {
		return nil
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetGrpcTimeout())
	defer cancel()

	logicalDevices, err := client.ListLogicalDevices(ctx, &empty.Empty{})
	if err != nil {
		return nil
	}

	list := make([]flags.Completion, 0)
	for _, item := range logicalDevices.Items {
		if strings.HasPrefix(item.Id, match) {
			list = append(list, flags.Completion{Item: item.Id})
		}
	}

	return list
}

func (options *LogicalDeviceList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetGrpcTimeout())
	defer cancel()

	logicalDevices, err := client.ListLogicalDevices(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	// Make sure json output prints an empty list, not "null"
	if logicalDevices.Items == nil {
		logicalDevices.Items = make([]*voltha.LogicalDevice, 0)
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("logical-device-list", "format", DEFAULT_LOGICAL_DEVICE_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("local-device-list", "order", "")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      logicalDevices.Items,
	}

	GenerateOutput(&result)
	return nil
}

func (options *LogicalDevicePortList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetGrpcTimeout())
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	ports, err := client.ListLogicalDevicePorts(ctx, &id)
	if err != nil {
		return err
	}

	// ensure no nil pointers
	for _, v := range ports.Items {
		if v.OfpPortStats == nil {
			v.OfpPortStats = &openflow_13.OfpPortStats{}
		}
		if v.OfpPort == nil {
			v.OfpPort = &openflow_13.OfpPort{}
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("logical-device-ports", "format", DEFAULT_LOGICAL_DEVICE_PORT_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("logical-device-ports", "order", "")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      ports.Items,
	}

	GenerateOutput(&result)
	return nil
}

func (options *LogicalDeviceFlowList) Execute(args []string) error {
	fl := &FlowList{}
	fl.ListOutputOptions = options.ListOutputOptions
	fl.FlowIdOptions = options.FlowIdOptions
	fl.Args.Id = string(options.Args.Id)
	fl.Method = "logical-device-flows"
	return fl.Execute(args)
}

func (options *LogicalDeviceInspect) Execute(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("only a single argument 'DEVICE_ID' can be provided")
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.GetGrpcTimeout())
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	logicalDevice, err := client.GetLogicalDevice(ctx, &id)
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("logical-device-inspect", "format", DEFAULT_LOGICAL_DEVICE_INSPECT_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      logicalDevice,
	}
	GenerateOutput(&result)
	return nil
}
