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
	"github.com/ciena/voltctl/pkg/format"
	"github.com/ciena/voltctl/pkg/model"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
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
	parser.AddCommand("devicegroup", "device group commands", "Commands to query and manipulate VOLTHA device groups",
		&deviceGroupOpts)
}

func (options *DeviceGroupList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("devicegroup-list")
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
		outputFormat = DEFAULT_DEVICE_GROUP_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	data := make([]model.DeviceGroup, len(items.([]interface{})))
	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].PopulateFrom(val)
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
