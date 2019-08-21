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
	"github.com/jhump/protoreflect/dynamic"
)

type PeerPort struct {
	DeviceId string `json:"deviceid"`
	PortNo   uint32 `json:"portno"`
}

type ProxyAddress struct {
	DeviceId           string `json:"deviceId"`
	DeviceType         string `json:"devicetype,omitempty"`
	ChannelId          uint32 `json:"channelid"`
	ChannelGroupId     uint32 `json:"channelgroup"`
	ChannelTermination string `json:"channeltermination,omitempty"`
	OnuId              uint32 `json:"onuid"`
	OnuSessionId       uint32 `json:"onusessionid"`
}

type Device struct {
	Id              string        `json:"id"`
	Type            string        `json:"type"`
	Root            bool          `json:"root"`
	ParentId        string        `json:"parentid"`
	ParentPortNo    uint32        `json:"parentportno"`
	Vendor          string        `json:"vendor"`
	Model           string        `json:"model"`
	HardwareVersion string        `json:"hardwareversion"`
	FirmwareVersion string        `json:"firmwareversion"`
	SerialNumber    string        `json:"serialnumber"`
	VendorId        string        `json:"vendorid"`
	Adapter         string        `json:"adapter"`
	Vlan            uint32        `json:"vlan"`
	MacAddress      string        `json:"macaddress"`
	Address         string        `json:"address"`
	ExtraArgs       string        `json:"extraargs"`
	ProxyAddress    *ProxyAddress `json:"proxyaddress,omitempty"`
	AdminState      string        `json:"adminstate"`
	OperStatus      string        `json:"operstatus"`
	Reason          string        `json:"reason"`
	ConnectStatus   string        `json:"connectstatus"`
	Ports           []DevicePort  `json:"ports"`
	Flows           []Flow        `json:"flows"`
}

type DevicePort struct {
	PortNo     uint32     `json:"portno"`
	Label      string     `json:"label"`
	Type       string     `json:"type"`
	AdminState string     `json:"adminstate"`
	OperStatus string     `json:"operstatus"`
	DeviceId   string     `json:"deviceid"`
	Peers      []PeerPort `json:"peers"`
}

func (d *Device) PopulateFrom(val *dynamic.Message) {
	d.Id = val.GetFieldByName("id").(string)
	d.Type = val.GetFieldByName("type").(string)
	d.Root = val.GetFieldByName("root").(bool)
	d.ParentId = val.GetFieldByName("parent_id").(string)
	d.ParentPortNo = val.GetFieldByName("parent_port_no").(uint32)
	d.Vendor = val.GetFieldByName("vendor").(string)
	d.Model = val.GetFieldByName("model").(string)
	d.HardwareVersion = val.GetFieldByName("hardware_version").(string)
	d.FirmwareVersion = val.GetFieldByName("firmware_version").(string)
	d.SerialNumber = val.GetFieldByName("serial_number").(string)
	d.VendorId = val.GetFieldByName("vendor_id").(string)
	d.Adapter = val.GetFieldByName("adapter").(string)
	d.MacAddress = val.GetFieldByName("mac_address").(string)
	d.Vlan = val.GetFieldByName("vlan").(uint32)
	d.Address = val.GetFieldByName("host_and_port").(string)
	if len(d.Address) == 0 {
		d.Address = val.GetFieldByName("ipv4_address").(string)
	}
	if len(d.Address) == 0 {
		d.Address = val.GetFieldByName("ipv6_address").(string)
	}
	if len(d.Address) == 0 {
		d.Address = "unknown"
	}
	d.ExtraArgs = val.GetFieldByName("extra_args").(string)
	proxy := val.GetFieldByName("proxy_address").(*dynamic.Message)
	d.ProxyAddress = nil
	if proxy != nil {
		d.ProxyAddress = &ProxyAddress{
			DeviceId:       proxy.GetFieldByName("device_id").(string),
			ChannelId:      proxy.GetFieldByName("channel_id").(uint32),
			ChannelGroupId: proxy.GetFieldByName("channel_group_id").(uint32),
			OnuId:          proxy.GetFieldByName("onu_id").(uint32),
			OnuSessionId:   proxy.GetFieldByName("onu_session_id").(uint32),
		}
		v, err := proxy.TryGetFieldByName("device_type")
		if err == nil {
			d.ProxyAddress.DeviceType = v.(string)
		}
		v, err = proxy.TryGetFieldByName("channel_termination")
		if err == nil {
			d.ProxyAddress.ChannelTermination = v.(string)
		}
	}
	d.AdminState = GetEnumValue(val, "admin_state")
	d.OperStatus = GetEnumValue(val, "oper_status")
	d.Reason = val.GetFieldByName("reason").(string)
	d.ConnectStatus = GetEnumValue(val, "connect_status")

	ports := val.GetFieldByName("ports").([]interface{})
	d.Ports = make([]DevicePort, len(ports))
	for i, port := range ports {
		d.Ports[i].PopulateFrom(port.(*dynamic.Message))
	}
	flows := val.GetFieldByName("flows").(*dynamic.Message)
	if flows == nil {
		d.Flows = make([]Flow, 0)
	} else {
		items := flows.GetFieldByName("items").([]interface{})
		d.Flows = make([]Flow, len(items))
		for i, flow := range items {
			d.Flows[i].PopulateFrom(flow.(*dynamic.Message))
		}
	}
}

func (port *DevicePort) PopulateFrom(val *dynamic.Message) {
	port.PortNo = val.GetFieldByName("port_no").(uint32)
	port.Type = GetEnumValue(val, "type")
	port.Label = val.GetFieldByName("label").(string)
	port.AdminState = GetEnumValue(val, "admin_state")
	port.OperStatus = GetEnumValue(val, "oper_status")
	port.DeviceId = val.GetFieldByName("device_id").(string)
	peers := val.GetFieldByName("peers").([]interface{})
	port.Peers = make([]PeerPort, len(peers))
	for j, peer := range peers {
		p := peer.(*dynamic.Message)
		port.Peers[j].DeviceId = p.GetFieldByName("device_id").(string)
		port.Peers[j].PortNo = p.GetFieldByName("port_no").(uint32)
	}
}
