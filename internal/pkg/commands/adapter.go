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
	DEFAULT_OUTPUT_FORMAT = "table{{ .Id }}\t{{.Vendor}}\t{{.Version}}"
)

type AdapterList struct {
	ListOutputOptions
}

type AdapterOpts struct {
	List AdapterList `command:"list"`
}

var adapterOpts = AdapterOpts{}

func RegisterAdapterCommands(parent *flags.Parser) {
	parent.AddCommand("adapter", "adapter commands", "Commands to query and manipulate VOLTHA adapters", &adapterOpts)
}

func (options *AdapterList) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("adapter-list")
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
		outputFormat = DEFAULT_OUTPUT_FORMAT
	}

	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	data := make([]model.Adapter, len(items.([]interface{})))
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
