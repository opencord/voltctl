/*
 * Copyright 2021-present Ciena Corporation
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
	"github.com/opencord/voltha-protos/v4/go/openflow_13"
	"github.com/opencord/voltha-protos/v4/go/voltha"
)

const (
	DEFAULT_DEVICE_GROUPS_FORMAT        = "table{{.GroupId}}\t{{.Type}}"
	DEFAULT_DEVICE_GROUPS_BUCKET_FORMAT = "table{{.GroupId}}\t{{.Buckets}}\t{{.Type}}"
)

type GroupList struct {
	ListOutputOptions
	GroupListOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`

	Method string
}

func (options *GroupList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Current().Grpc.Timeout)
	defer cancel()

	id := voltha.ID{Id: string(options.Args.Id)}

	var groups *openflow_13.FlowGroups
	switch options.Method {
	case "device-groups":
		groups, err = client.ListDeviceFlowGroups(ctx, &id)
	case "logical-device-groups":
		groups, err = client.ListLogicalDeviceFlowGroups(ctx, &id)
	default:
		Error.Fatalf("Unknown method name: '%s'", options.Method)
	}

	var groupList []*openflow_13.OfpGroupDesc
	if err != nil {
		return err
	}
	if toOutputType(options.OutputAs) == OUTPUT_TABLE && (groups == nil || len(groups.Items) == 0) {
		fmt.Println("*** NO GROUPS ***")
		return nil
	}
	for _, item := range groups.Items {
		if item.Desc.Type != openflow_13.OfpGroupType_OFPGT_FF {
			// Since onos is setting watch port and watch group as max uint32,
			// for group type other than FF, we need to remove watch port and group for them.
			for _, bucket := range item.Desc.Buckets {
				bucket.WatchPort = 0
				bucket.WatchGroup = 0
			}
		}
		groupList = append(groupList, item.Desc)
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if options.Quiet {
		outputFormat = "{{.GroupId}}"
	} else if outputFormat == "" {
		if options.Bucket {
			outputFormat = GetCommandOptionWithDefault(options.Method, "format", DEFAULT_DEVICE_GROUPS_BUCKET_FORMAT)
		} else {
			outputFormat = GetCommandOptionWithDefault(options.Method, "format", DEFAULT_DEVICE_GROUPS_FORMAT)
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
		Data:      groupList,
	}
	GenerateOutput(&result)

	return nil
}
