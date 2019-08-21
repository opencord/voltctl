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
package model

import (
	"fmt"
	"github.com/jhump/protoreflect/dynamic"
	"strings"
)

type LogicalDevice struct {
	Id           string `json:"id"`
	DatapathId   string `json:"datapathid"`
	RootDeviceId string `json:"rootdeviceid"`
	SerialNumber string `json:"serialnumber"`
	Features     struct {
		NBuffers     uint32 `json:"nbuffers"`
		NTables      uint32 `json:"ntables"`
		Capabilities string `json:"capabilities"`
	} `json:"features"`
	Ports []LogicalPort `json:"ports"`
	Flows []Flow        `json:"flows"`
}

type LogicalPort struct {
	Id           string `json:"id"`
	DeviceId     string `json:"deviceid"`
	DevicePortNo uint32 `json:"deviceportno"`
	RootPort     bool   `json:"rootport"`
	Openflow     struct {
		PortNo   uint32 `json:"portno"`
		HwAddr   string `json:"hwaddr"`
		Name     string `json:"name"`
		Config   string `json:"config"`
		State    string `json:"state"`
		Features struct {
			Advertised string `json:"advertised"`
			Current    string `json:"current"`
			Supported  string `json:"supported"`
			Peer       string `json:"peer"`
		} `json:"features"`
		Bitrate struct {
			Current uint32 `json:"current"`
			Max     uint32 `json:"max"`
		}
	} `json:"openflow"`
}

func (device *LogicalDevice) PopulateFrom(val *dynamic.Message) {
	device.Id = val.GetFieldByName("id").(string)
	device.DatapathId = fmt.Sprintf("%016x", val.GetFieldByName("datapath_id").(uint64))
	device.RootDeviceId = val.GetFieldByName("root_device_id").(string)
	desc := val.GetFieldByName("desc").(*dynamic.Message)
	device.SerialNumber = desc.GetFieldByName("serial_num").(string)
	features := val.GetFieldByName("switch_features").(*dynamic.Message)
	device.Features.NBuffers = features.GetFieldByName("n_buffers").(uint32)
	device.Features.NTables = features.GetFieldByName("n_tables").(uint32)
	device.Features.Capabilities = fmt.Sprintf("0x%08x", features.GetFieldByName("capabilities").(uint32))

	ports := val.GetFieldByName("ports").([]interface{})
	device.Ports = make([]LogicalPort, len(ports))
	for i, port := range ports {
		device.Ports[i].PopulateFrom(port.(*dynamic.Message))
	}

	flows := val.GetFieldByName("flows").(*dynamic.Message)
	if flows == nil {
		device.Flows = make([]Flow, 0)
	} else {
		items := flows.GetFieldByName("items").([]interface{})
		device.Flows = make([]Flow, len(items))
		for i, flow := range items {
			device.Flows[i].PopulateFrom(flow.(*dynamic.Message))
		}
	}
}

func (port *LogicalPort) PopulateFrom(val *dynamic.Message) {
	port.Id = val.GetFieldByName("id").(string)
	port.DeviceId = val.GetFieldByName("device_id").(string)
	port.DevicePortNo = val.GetFieldByName("device_port_no").(uint32)
	port.RootPort = val.GetFieldByName("root_port").(bool)
	ofp := val.GetFieldByName("ofp_port").(*dynamic.Message)
	hw := strings.Builder{}
	first := true
	for _, b := range ofp.GetFieldByName("hw_addr").([]interface{}) {
		if !first {
			hw.WriteString(":")
		}
		first = false
		hw.WriteString(fmt.Sprintf("%02x", b))
	}
	port.Openflow.HwAddr = hw.String()
	port.Openflow.PortNo = ofp.GetFieldByName("port_no").(uint32)
	port.Openflow.Name = ofp.GetFieldByName("name").(string)
	port.Openflow.Config = fmt.Sprintf("0x%08x", ofp.GetFieldByName("config").(uint32))
	port.Openflow.State = fmt.Sprintf("0x%08x", ofp.GetFieldByName("state").(uint32))
	port.Openflow.Features.Current = fmt.Sprintf("0x%08x", ofp.GetFieldByName("curr").(uint32))
	port.Openflow.Features.Advertised = fmt.Sprintf("0x%08x", ofp.GetFieldByName("advertised").(uint32))
	port.Openflow.Features.Supported = fmt.Sprintf("0x%08x", ofp.GetFieldByName("supported").(uint32))
	port.Openflow.Features.Peer = fmt.Sprintf("0x%08x", ofp.GetFieldByName("peer").(uint32))
	port.Openflow.Bitrate.Current = ofp.GetFieldByName("curr_speed").(uint32)
	port.Openflow.Bitrate.Max = ofp.GetFieldByName("max_speed").(uint32)
}
