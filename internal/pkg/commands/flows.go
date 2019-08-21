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
	"github.com/jhump/protoreflect/dynamic"
	"sort"
	"strings"
)

type FlowList struct {
	ListOutputOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`

	Method string
}

type FlowOpts struct {
	List FlowList `command:"list"`
}

var (
	flowOpts = FlowOpts{}

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
		"ClearActions":           1020,
	}
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

func toVlanId(vid uint32) string {
	if vid == 0 {
		return "untagged"
	} else if vid&0x1000 > 0 {
		return fmt.Sprintf("%d", vid-4096)
	}
	return fmt.Sprintf("%d", vid)
}

func appendInt32(base string, val int32) string {
	if len(base) > 0 {
		return fmt.Sprintf("%s,%d", base, val)
	}
	return fmt.Sprintf("%d", val)
}

func appendUint32(base string, val uint32) string {
	if len(base) > 0 {
		return fmt.Sprintf("%s,%d", base, val)
	}
	return fmt.Sprintf("%d", val)
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

	switch options.Method {
	case "device-flow-list":
	case "logical-device-flow-list":
	default:
		panic(fmt.Errorf("Unknown method name: '%s'", options.Method))
	}

	descriptor, method, err := GetMethod(options.Method)
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
	} else if h.Status != nil && h.Status.Err() != nil {
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

	if toOutputType(options.OutputAs) == OUTPUT_TABLE && (items == nil || len(items.([]interface{})) == 0) {
		fmt.Println("*** NO FLOWS ***")
		return nil
	}

	// Walk the flows and populate the output table
	data := make([]model.Flow, len(items.([]interface{})))
	var fieldset model.FlowFieldFlag
	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].PopulateFrom(val)
		fieldset |= data[i].Populated()
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if options.Quiet {
		outputFormat = "{{.Id}}"
	} else if outputFormat == "" {
		outputFormat = buildOutputFormat(fieldset, model.FLOW_FIELD_STATS)
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
