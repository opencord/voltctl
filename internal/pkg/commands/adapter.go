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
	"github.com/jessevdk/go-flags"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltha-protos/v4/go/voltha"
)

const (
	DEFAULT_OUTPUT_FORMAT = "table{{ .Id }}\t{{ .Vendor }}\t{{ .Type }}\t{{ .Endpoint }}\t{{ .Version }}\t{{ .CurrentReplica }}\t{{ .TotalReplicas }}\t{{ since .LastCommunication}}"
)

type AdapterList struct {
	ListOutputOptions
}

type AdapterOpts struct {
	List AdapterList `command:"list"`
}

var adapterOpts = AdapterOpts{}

func RegisterAdapterCommands(parent *flags.Parser) {
	if _, err := parent.AddCommand("adapter", "adapter commands", "Commands to query and manipulate VOLTHA adapters", &adapterOpts); err != nil {
		Error.Fatalf("Unexpected error while attempting to register adapter commands : %s", err)
	}
}

func (options *AdapterList) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	adapters, err := client.ListAdapters(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault("adapter-list", "format", DEFAULT_OUTPUT_FORMAT)
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}
	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault("adapter-list", "order", "")
	}

	// TODO: lastCommunication ends up formatted as `seconds:1589415656 nanos:775740000`
	//   need to think through where to do presentation formatting.

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      adapters.Items,
	}
	GenerateOutput(&result)

	return nil
}
