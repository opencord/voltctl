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
	"github.com/opencord/voltha-protos/v4/go/openflow_13"
	"github.com/opencord/voltha-protos/v4/go/voltha"
	//"sort"
	"strings"
)

const (
	//DEFAULT_DEVICE_FLOW_GROUPS_FORMAT = "table{{desc.GroupId}}\t{{Desc.Type}}\t{{Stats.RefCount}}\t{{Stats.PacketCount}}\t{{Stats.ByteCount}}"
	DEFAULT_DEVICE_FLOW_GROUPS_FORMAT = "table{{.GroupId}}\t{{.GroupType}}\t{{.RefCount}}\t{{.PacketCount}}\t{{.ByteCount}}"
	DEFAULT_DEVICE_FLOW_GROUPS_BUCKET_FORMAT = "table{{.GroupId}}\t{{.GroupType}}\t{{.RefCount}}\t{{.PacketCount}}\t{{.ByteCount}}\t{{.Buckets}}"
)

type GroupList struct {
	ListOutputOptions
	GroupListOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`

	Method string
}

type GroupOpts struct {
	List GroupList `command:"list"`
}

/*
var (
	// Used to sort the table colums in a consistent order
	SORT_GROUP_ORDER = map[string]uint16{
		"GroupId":                     0,
		"GroupType":                  10,
		"DurationSec":               20,
		"DurationNsec":                 30,
		"RefCount":       35,
		"PacketCount":                 40,
		"ByteCount":                 50,
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
)

/*
 * Construct a template format string based on the fields required by the
 * results.
*/
func buildGroupOutputFormat(fieldset model.GroupFieldFlag) string {
	want := fieldset
	fmt.Println(want)
	fields := make([]string, want.Count())
	idx := 0
	for _, flag := range model.AllGroupFieldFlags {
		if want.IsSet(flag) {
			fields[idx] = flag.GetFormatString()
			idx += 1
		}
	}
	/*
	sort.Slice(fields, func(i, j int) bool {
		return SORT_ORDER[fields[i]] < SORT_ORDER[fields[j]]
	})

	 */
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
	fmt.Println(b.String())
	return b.String()
}

func (options *GroupList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	var groups *openflow_13.FlowGroups
	fmt.Println(options)
	switch options.Method {
	case "device-groups":
		groups, err = client.ListDeviceFlowGroups(ctx, &id)
	case "logical-device-groups":
		groups, err = client.ListLogicalDeviceFlowGroups(ctx, &id)
	default:
		Error.Fatalf("Unknown method name: '%s'", options.Method)
	}

	if err != nil {
		return err
	}
	fmt.Println(groups == nil, groups.Items)
	if toOutputType(options.OutputAs) == OUTPUT_TABLE && (groups == nil || len(groups.Items) == 0) {
		fmt.Println("*** NO GROUPS ***")
		return nil
	}
	fmt.Println(groups)
	data := make([]model.Group, len(groups.Items))
	var fieldset model.GroupFieldFlag
	for i, item := range groups.Items {
		data[i].PopulateGroupFormatFromProto(item)
		fieldset |= data[i].Populated()
	}
	outputFormat := CharReplacer.Replace(options.Format)
	if options.Quiet {
		outputFormat = "{{.GroupId}}"
	} else if outputFormat == "" {
		if options.Bucket {
			outputFormat = GetCommandOptionWithDefault(options.Method, "format", DEFAULT_DEVICE_FLOW_GROUPS_BUCKET_FORMAT)
		} else {
			outputFormat = GetCommandOptionWithDefault(options.Method, "format", DEFAULT_DEVICE_FLOW_GROUPS_FORMAT)
		}
	}

	orderBy := options.OrderBy
	if orderBy == "" {
		orderBy = GetCommandOptionWithDefault(options.Method, "order", "")
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
