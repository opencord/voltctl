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
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/model"
	"github.com/opencord/voltha-protos/v4/go/voltha"
)

type ReasonList struct {
	ListOutputOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`

	Method string
}

func (options *ReasonList) Execute(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("only a single argument 'DEVICE_ID' can be provided")
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	reasons, err := client.GetDeviceReasons(ctx, &id)

	if err != nil {
		return err
	}

	if toOutputType(options.OutputAs) == OUTPUT_TABLE && (reasons == nil || len(reasons.Items) == 0) {
		fmt.Println("*** NO REASONS AVAILABLE ***")
		return nil
	}

	data := make([]model.Reason, len(reasons.Items))
	for i, item := range reasons.Items {
		data[i].PopulateFromProto(item)
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if options.Quiet {
		outputFormat = "{{.Reason}}"
	} else if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault(options.Method, "format", DEFAULT_DEVICE_REASON_FORMAT)
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault(options.Method, "order", "Reason")
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   orderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}
	GenerateOutput(&result)

	return nil
}
