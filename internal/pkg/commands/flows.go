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
	"github.com/opencord/voltha-protos/v5/go/openflow_13"
	"github.com/opencord/voltha-protos/v5/go/voltha"
	"sort"
	"strings"
)

type FlowList struct {
	ListOutputOptions
	FlowIdOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`

	Method string
}

type FlowOpts struct {
	List FlowList `command:"list"`
}

var (
	// Used to sort the table colums in a consistent order
	SORT_ORDER = map[string]uint16{
		"Id":                     0,
		"TableId":                10,
		"Priority":               20,
		"Cookie":                 30,
		"UnsupportedMatch":       35,
		"InPort":                 40,
		"VlanId":                 50,
		"VlanPcp":                55,
		"EthType":                60,
		"IpProto":                70,
		"UdpSrc":                 80,
		"UdpDst":                 90,
		"Metadata":               100,
		"TunnelId":               101,
		"UnsupportedInstruction": 102,
		"UnsupportedAction":      105,
		"UnsupportedSetField":    107,
		"SetVlanId":              110,
		"PopVlan":                120,
		"PushVlanId":             130,
		"Output":                 1000,
		"GotoTable":              1010,
		"WriteMetadata":          1015,
		"ClearActions":           1020,
		"MeterId":                1030,
	}
	DEFAULT_FLOWS_ORDER = "Id"
)

/*
 * Construct a template format string based on the fields required by the
 * results.
 */
func buildOutputFormat(fieldset model.FlowFieldFlag, ignore model.FlowFieldFlag) string {
	want := fieldset & ^(ignore)
	fields := make([]string, want.Count())
	idx := 0
	for _, flag := range model.AllFlowFieldFlags {
		if want.IsSet(flag) {
			fields[idx] = flag.String()
			idx += 1
		}
	}
	sort.Slice(fields, func(i, j int) bool {
		return SORT_ORDER[fields[i]] < SORT_ORDER[fields[j]]
	})
	var b strings.Builder
	b.WriteString("table")
	first := true
	for _, k := range fields {
		if !first {
			b.WriteString("\t")
		}
		first = false
		b.WriteString("{{.")
		b.WriteString(k)
		b.WriteString("}}")
	}
	return b.String()
}

func (options *FlowList) Execute(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("only a single argument 'DEVICE_ID' can be provided")
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	var flows *openflow_13.Flows

	switch options.Method {
	case "device-flows":
		flows, err = client.ListDeviceFlows(ctx, &id)
	case "logical-device-flows":
		flows, err = client.ListLogicalDeviceFlows(ctx, &id)
	default:
		Error.Fatalf("Unknown method name: '%s'", options.Method)
	}

	if err != nil {
		return err
	}

	if toOutputType(options.OutputAs) == OUTPUT_TABLE && (flows == nil || len(flows.Items) == 0) {
		fmt.Println("*** NO FLOWS ***")
		return nil
	}

	data := make([]model.Flow, len(flows.Items))
	var fieldset model.FlowFieldFlag
	for i, item := range flows.Items {
		data[i].PopulateFromProto(item, options.HexId)
		fieldset |= data[i].Populated()
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if options.Quiet {
		outputFormat = "{{.Id}}"
	} else if outputFormat == "" {
		outputFormat = GetCommandOptionWithDefault(options.Method, "format", buildOutputFormat(fieldset, model.FLOW_FIELD_STATS))
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault(options.Method, "order", DEFAULT_FLOWS_ORDER)
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
