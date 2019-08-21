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
	"github.com/ciena/voltctl/pkg/format"
	"github.com/ciena/voltctl/pkg/model"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"strings"
)

const (
	DEFAULT_LOGICAL_DEVICE_FORMAT         = "table{{ .Id }}\t{{.DatapathId}}\t{{.RootDeviceId}}\t{{.SerialNumber}}\t{{.Features.NBuffers}}\t{{.Features.NTables}}\t{{.Features.Capabilities}}"
	DEFAULT_LOGICAL_DEVICE_PORT_FORMAT    = "table{{.Id}}\t{{.DeviceId}}\t{{.DevicePortNo}}\t{{.RootPort}}\t{{.Openflow.PortNo}}\t{{.Openflow.HwAddr}}\t{{.Openflow.Name}}\t{{.Openflow.State}}\t{{.Openflow.Features.Current}}\t{{.Openflow.Bitrate.Current}}"
	DEFAULT_LOGICAL_DEVICE_INSPECT_FORMAT = `ID: {{.Id}}
  DATAPATHID: {{.DatapathId}}
  ROOTDEVICEID: {{.RootDeviceId}}
  SERIALNUMNER: {{.SerialNumber}}`
)

type LogicalDeviceId string

type LogicalDeviceList struct {
	ListOutputOptions
}

type LogicalDeviceFlowList struct {
	ListOutputOptions
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
	List    LogicalDeviceList     `command:"list"`
	Flows   LogicalDeviceFlowList `command:"flows"`
	Ports   LogicalDevicePortList `command:"ports"`
	Inspect LogicalDeviceInspect  `command:"inspect"`
}

var logicalDeviceOpts = LogicalDeviceOpts{}

func RegisterLogicalDeviceCommands(parser *flags.Parser) {
	parser.AddCommand("logicaldevice", "logical device commands", "Commands to query and manipulate VOLTHA logical devices", &logicalDeviceOpts)
}

func (i *LogicalDeviceId) Complete(match string) []flags.Completion {
	conn, err := NewConnection()
	if err != nil {
		return nil
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("logical-device-list")
	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
	if err != nil {
		return nil
	}

	if h.Status != nil && h.Status.Err() != nil {
		return nil
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return nil
	}

	items, err := d.TryGetFieldByName("items")
	if err != nil {
		return nil
	}

	list := make([]flags.Completion, 0)
	for _, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		id := val.GetFieldByName("id").(string)
		if strings.HasPrefix(id, match) {
			list = append(list, flags.Completion{Item: id})
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

	descriptor, method, err := GetMethod("logical-device-list")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
	if err != nil {
		return err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	items, err := d.TryGetFieldByName("items")
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_LOGICAL_DEVICE_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	data := make([]model.LogicalDevice, len(items.([]interface{})))
	for i, item := range items.([]interface{}) {
		data[i].PopulateFrom(item.(*dynamic.Message))
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   options.OrderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
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

	descriptor, method, err := GetMethod("logical-device-ports")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{ParamNames[GlobalConfig.ApiVersion]["ID"]: {"id": options.Args.Id}},
	}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
	if err != nil {
		return err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	items, err := d.TryGetFieldByName("items")
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_LOGICAL_DEVICE_PORT_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	data := make([]model.LogicalPort, len(items.([]interface{})))
	for i, item := range items.([]interface{}) {
		data[i].PopulateFrom(item.(*dynamic.Message))
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   options.OrderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}

	GenerateOutput(&result)
	return nil
}

func (options *LogicalDeviceFlowList) Execute(args []string) error {
	fl := &FlowList{}
	fl.ListOutputOptions = options.ListOutputOptions
	fl.Args.Id = string(options.Args.Id)
	fl.Method = "logical-device-flow-list"
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

	descriptor, method, err := GetMethod("logical-device-inspect")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{ParamNames[GlobalConfig.ApiVersion]["ID"]: {"id": options.Args.Id}},
	}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{"Get-Depth: 2"}, h, h.GetParams)
	if err != nil {
		return err
	} else if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}

	device := &model.LogicalDevice{}
	device.PopulateFrom(d)

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_LOGICAL_DEVICE_INSPECT_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      device,
	}
	GenerateOutput(&result)
	return nil
}
