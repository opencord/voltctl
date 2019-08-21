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
)

type FlowFieldFlag uint64

const (

	// Define bit flags for flow fields to determine what is set and
	// what is not
	FLOW_FIELD_UNSUPPORTED_MATCH FlowFieldFlag = 1 << iota
	FLOW_FIELD_UNSUPPORTED_INSTRUCTION
	FLOW_FIELD_UNSUPPORTED_ACTION
	FLOW_FIELD_UNSUPPORTED_SET_FIELD
	FLOW_FIELD_ID
	FLOW_FIELD_TABLE_ID
	FLOW_FIELD_DURATION_SEC
	FLOW_FIELD_DURATION_NSEC
	FLOW_FIELD_IDLE_TIMEOUT
	FLOW_FIELD_HARD_TIMEOUT
	FLOW_FIELD_PACKET_COUNT
	FLOW_FIELD_BYTE_COUNT
	FLOW_FIELD_PRIORITY
	FLOW_FIELD_COOKIE
	FLOW_FIELD_IN_PORT
	FLOW_FIELD_ETH_TYPE
	FLOW_FIELD_VLAN_ID
	FLOW_FIELD_IP_PROTO
	FLOW_FIELD_UDP_SRC
	FLOW_FIELD_UDP_DST
	FLOW_FIELD_METADATA
	FLOW_FIELD_SET_VLAN_ID
	FLOW_FIELD_POP_VLAN
	FLOW_FIELD_PUSH_VLAN_ID
	FLOW_FIELD_OUTPUT
	FLOW_FIELD_GOTO_TABLE
	FLOW_FIELD_CLEAR_ACTIONS
	FLOW_FIELD_TUNNEL_ID
	FLOW_FIELD_VLAN_PCP

	FLOW_FIELD_HEADER = FLOW_FIELD_ID | FLOW_FIELD_TABLE_ID |
		FLOW_FIELD_PRIORITY | FLOW_FIELD_COOKIE

	FLOW_FIELD_STATS = FLOW_FIELD_DURATION_SEC | FLOW_FIELD_DURATION_NSEC |
		FLOW_FIELD_IDLE_TIMEOUT | FLOW_FIELD_HARD_TIMEOUT |
		FLOW_FIELD_PACKET_COUNT | FLOW_FIELD_BYTE_COUNT
)

var (
	// Provide an array of all flags that can be used for iteration
	AllFlowFieldFlags = []FlowFieldFlag{
		FLOW_FIELD_UNSUPPORTED_MATCH,
		FLOW_FIELD_UNSUPPORTED_INSTRUCTION,
		FLOW_FIELD_UNSUPPORTED_ACTION,
		FLOW_FIELD_UNSUPPORTED_SET_FIELD,
		FLOW_FIELD_ID,
		FLOW_FIELD_TABLE_ID,
		FLOW_FIELD_DURATION_SEC,
		FLOW_FIELD_DURATION_NSEC,
		FLOW_FIELD_IDLE_TIMEOUT,
		FLOW_FIELD_HARD_TIMEOUT,
		FLOW_FIELD_PACKET_COUNT,
		FLOW_FIELD_BYTE_COUNT,
		FLOW_FIELD_PRIORITY,
		FLOW_FIELD_COOKIE,
		FLOW_FIELD_IN_PORT,
		FLOW_FIELD_ETH_TYPE,
		FLOW_FIELD_VLAN_ID,
		FLOW_FIELD_IP_PROTO,
		FLOW_FIELD_UDP_SRC,
		FLOW_FIELD_UDP_DST,
		FLOW_FIELD_METADATA,
		FLOW_FIELD_SET_VLAN_ID,
		FLOW_FIELD_POP_VLAN,
		FLOW_FIELD_PUSH_VLAN_ID,
		FLOW_FIELD_OUTPUT,
		FLOW_FIELD_GOTO_TABLE,
		FLOW_FIELD_CLEAR_ACTIONS,
		FLOW_FIELD_TUNNEL_ID,
		FLOW_FIELD_VLAN_PCP,
	}
)

func (f *FlowFieldFlag) Count() int {
	var count int
	var bit uint64 = 1
	var asUint64 = uint64(*f)
	for i := 0; i < 64; i += 1 {
		if asUint64&bit > 0 {
			count += 1
		}
		bit <<= 1
	}
	return count
}
func (f *FlowFieldFlag) IsSet(flag FlowFieldFlag) bool {
	return *f&flag > 0
}

func (f *FlowFieldFlag) Set(flag FlowFieldFlag) {
	*f |= flag
}

func (f *FlowFieldFlag) Clear(flag FlowFieldFlag) {
	var mask = ^(flag)
	*f &= mask
}

func (f *FlowFieldFlag) Reset() {
	*f = 0
}

func (f FlowFieldFlag) String() string {
	switch f {
	case FLOW_FIELD_UNSUPPORTED_MATCH:
		return "UnsupportedMatch"
	case FLOW_FIELD_UNSUPPORTED_INSTRUCTION:
		return "UnsupportedInstruction"
	case FLOW_FIELD_UNSUPPORTED_ACTION:
		return "UnsupportedAction"
	case FLOW_FIELD_UNSUPPORTED_SET_FIELD:
		return "UnsupportedSetField"
	case FLOW_FIELD_ID:
		return "Id"
	case FLOW_FIELD_TABLE_ID:
		return "TableId"
	case FLOW_FIELD_DURATION_SEC:
		return "DurationSec"
	case FLOW_FIELD_DURATION_NSEC:
		return "DurationNsec"
	case FLOW_FIELD_IDLE_TIMEOUT:
		return "IdleTimeout"
	case FLOW_FIELD_HARD_TIMEOUT:
		return "HardTimeout"
	case FLOW_FIELD_PACKET_COUNT:
		return "PacketCount"
	case FLOW_FIELD_BYTE_COUNT:
		return "ByteCount"
	case FLOW_FIELD_PRIORITY:
		return "Priority"
	case FLOW_FIELD_COOKIE:
		return "Cookie"
	case FLOW_FIELD_IN_PORT:
		return "InPort"
	case FLOW_FIELD_ETH_TYPE:
		return "EthType"
	case FLOW_FIELD_VLAN_ID:
		return "VlanId"
	case FLOW_FIELD_IP_PROTO:
		return "IpProto"
	case FLOW_FIELD_UDP_SRC:
		return "UdpSrc"
	case FLOW_FIELD_UDP_DST:
		return "UdpDst"
	case FLOW_FIELD_METADATA:
		return "Metadata"
	case FLOW_FIELD_SET_VLAN_ID:
		return "SetVlanId"
	case FLOW_FIELD_POP_VLAN:
		return "PopVlan"
	case FLOW_FIELD_PUSH_VLAN_ID:
		return "PushVlanId"
	case FLOW_FIELD_OUTPUT:
		return "Output"
	case FLOW_FIELD_GOTO_TABLE:
		return "GotoTable"
	case FLOW_FIELD_CLEAR_ACTIONS:
		return "ClearActions"
	case FLOW_FIELD_TUNNEL_ID:
		return "TunnelId"
	case FLOW_FIELD_VLAN_PCP:
		return "VlanPcp"
	default:
		return "UnknownFieldFlag"
	}
}

/*
 * This is a partial list of OF match/action values. This list will be
 * expanded as new fields are needed within VOLTHA
 *
 * Strings are used in the output structure so that on output the table
 * can be "sparsely" populated with "empty" cells as opposed to 0 (zeros)
 * all over the place.
 */
type Flow struct {
	Id                     string `json:"id"`
	TableId                uint32 `json:"tableid"`
	DurationSec            uint32 `json:"durationsec"`
	DurationNsec           uint32 `json:"durationnsec"`
	IdleTimeout            uint32 `json:"idletimeout"`
	HardTimeout            uint32 `json:"hardtimeout"`
	PacketCount            uint64 `json:"packetcount"`
	ByteCount              uint64 `json:"bytecount"`
	Priority               uint32 `json:"priority"`
	Cookie                 string `json:"cookie"`
	UnsupportedMatch       string `json:"unsupportedmatch,omitempty"`
	InPort                 string `json:"inport,omitempty"`
	EthType                string `json:"ethtype,omitempty"`
	VlanId                 string `json:"vlanid,omitempty"`
	IpProto                string `json:"ipproto,omitempty"`
	UdpSrc                 string `json:"udpsrc,omitempty"`
	UdpDst                 string `json:"dstsrc,omitempty"`
	Metadata               string `json:"metadata,omitempty"`
	UnsupportedInstruction string `json:"unsupportedinstruction,omitempty"`
	UnsupportedAction      string `json:"unsupportedaction,omitempty"`
	UnsupportedSetField    string `json:"unsupportedsetfield,omitempty"`
	SetVlanId              string `json:"setvlanid,omitempty"`
	PopVlan                string `json:"popvlan,omitempty"`
	PushVlanId             string `json:"pushvlanid,omitempty"`
	Output                 string `json:"output,omitempty"`
	GotoTable              string `json:"gototable,omitempty"`
	ClearActions           string `json:"clear,omitempty"`
	TunnelId               string `json:"tunnelid,omitempty"`
	VlanPcp                string `json:"vlanpcp,omitempty"`

	populated FlowFieldFlag
}

func (f *Flow) Count() int {
	return f.populated.Count()
}

func (f *Flow) IsSet(flag FlowFieldFlag) bool {
	return f.populated.IsSet(flag)
}

func (f *Flow) Set(flag FlowFieldFlag) {
	f.populated.Set(flag)
}

func (f *Flow) Clear(flag FlowFieldFlag) {
	f.populated.Clear(flag)
}

func (f *Flow) Reset() {
	f.populated.Reset()
}

func (f *Flow) Populated() FlowFieldFlag {
	return f.populated
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

func (f *Flow) PopulateFrom(val *dynamic.Message) {

	f.Reset()
	f.Id = fmt.Sprintf("%016x", val.GetFieldByName("id").(uint64))
	f.TableId = val.GetFieldByName("table_id").(uint32)
	f.Priority = val.GetFieldByName("priority").(uint32)
	// mask the lower 8 for the cookie, why?
	cookie := val.GetFieldByName("cookie").(uint64)
	if cookie == 0 {
		f.Cookie = "0"
	} else {
		f.Cookie = fmt.Sprintf("~%08x", val.GetFieldByName("cookie").(uint64)&0xffffffff)
	}
	f.DurationSec = val.GetFieldByName("duration_sec").(uint32)
	f.DurationNsec = val.GetFieldByName("duration_nsec").(uint32)
	f.IdleTimeout = val.GetFieldByName("idle_timeout").(uint32)
	f.HardTimeout = val.GetFieldByName("hard_timeout").(uint32)
	f.PacketCount = val.GetFieldByName("packet_count").(uint64)
	f.ByteCount = val.GetFieldByName("byte_count").(uint64)
	f.Set(FLOW_FIELD_HEADER | FLOW_FIELD_STATS)

	match := val.GetFieldByName("match").(*dynamic.Message)
	fields := match.GetFieldByName("oxm_fields")
	for _, ifield := range fields.([]interface{}) {
		field := ifield.(*dynamic.Message)

		// Only support OFPXMC_OPENFLOW_BASIC (0x8000)
		if field.GetFieldByName("oxm_class").(int32) != 0x8000 {
			continue
		}

		basic := field.GetFieldByName("ofb_field").(*dynamic.Message)
		switch basic.GetFieldByName("type").(int32) {
		case 0: // IN_PORT
			f.Set(FLOW_FIELD_IN_PORT)
			f.InPort = fmt.Sprintf("%d", basic.GetFieldByName("port").(uint32))
		case 2: // METADATA
			f.Set(FLOW_FIELD_METADATA)
			f.Metadata = fmt.Sprintf("0x%016x", basic.GetFieldByName("table_metadata").(uint64))
		case 5: // ETH_TYPE
			f.Set(FLOW_FIELD_ETH_TYPE)
			f.EthType = fmt.Sprintf("0x%04x", basic.GetFieldByName("eth_type").(uint32))
		case 6: // VLAN_ID
			f.Set(FLOW_FIELD_VLAN_ID)
			f.VlanId = toVlanId(basic.GetFieldByName("vlan_vid").(uint32))
		case 7: // VLAN_PCP
			f.Set(FLOW_FIELD_VLAN_PCP)
			f.VlanPcp = fmt.Sprintf("%d", basic.GetFieldByName("vlan_pcp").(uint32))
		case 10: // IP_PROTO
			f.Set(FLOW_FIELD_IP_PROTO)
			f.IpProto = fmt.Sprintf("%d", basic.GetFieldByName("ip_proto").(uint32))
		case 15: // UDP_SRC
			f.Set(FLOW_FIELD_UDP_SRC)
			f.UdpSrc = fmt.Sprintf("%d", basic.GetFieldByName("udp_src").(uint32))
		case 16: // UDP_DST
			f.Set(FLOW_FIELD_UDP_DST)
			f.UdpDst = fmt.Sprintf("%d", basic.GetFieldByName("udp_dst").(uint32))
		case 38: // TUNNEL_ID
			f.Set(FLOW_FIELD_TUNNEL_ID)
			f.TunnelId = fmt.Sprintf("%d", basic.GetFieldByName("tunnel_id").(uint64))
		default:
			/*
			 * For unsupported match types put them into an
			 * "Unsupported field so the table/json still
			 * outputs relatively correctly as opposed to
			 * having log messages.
			 */
			f.Set(FLOW_FIELD_UNSUPPORTED_MATCH)
			f.UnsupportedMatch = appendInt32(f.UnsupportedMatch, basic.GetFieldByName("type").(int32))
		}
	}
	for _, instruction := range val.GetFieldByName("instructions").([]interface{}) {
		inst := instruction.(*dynamic.Message)
		switch inst.GetFieldByName("type").(uint32) {
		case 1: // GOTO_TABLE
			f.Set(FLOW_FIELD_GOTO_TABLE)
			goto_table := inst.GetFieldByName("goto_table").(*dynamic.Message)
			f.GotoTable = fmt.Sprintf("%d", goto_table.GetFieldByName("table_id").(uint32))
		case 4: // APPLY_ACTIONS
			actions := inst.GetFieldByName("actions").(*dynamic.Message)
			for _, action := range actions.GetFieldByName("actions").([]interface{}) {
				a := action.(*dynamic.Message)
				switch a.GetFieldByName("type").(int32) {
				case 0: // OUTPUT
					f.Set(FLOW_FIELD_OUTPUT)
					output := a.GetFieldByName("output").(*dynamic.Message)
					out := output.GetFieldByName("port").(uint32)
					switch out & 0x7fffffff {
					case 0:
						f.Output = "INVALID"
					case 0x7ffffff8:
						f.Output = "IN_PORT"
					case 0x7ffffff9:
						f.Output = "TABLE"
					case 0x7ffffffa:
						f.Output = "NORMAL"
					case 0x7ffffffb:
						f.Output = "FLOOD"
					case 0x7ffffffc:
						f.Output = "ALL"
					case 0x7ffffffd:
						f.Output = "CONTROLLER"
					case 0x7ffffffe:
						f.Output = "LOCAL"
					case 0x7fffffff:
						f.Output = "ANY"
					default:
						f.Output = fmt.Sprintf("%d", output.GetFieldByName("port").(uint32))
					}
				case 17: // PUSH_VLAN
					f.Set(FLOW_FIELD_PUSH_VLAN_ID)
					push := a.GetFieldByName("push").(*dynamic.Message)
					f.PushVlanId = fmt.Sprintf("0x%x", push.GetFieldByName("ethertype").(uint32))
				case 18: // POP_VLAN
					f.Set(FLOW_FIELD_POP_VLAN)
					f.PopVlan = "yes"
				case 25: // SET_FIELD
					set := a.GetFieldByName("set_field").(*dynamic.Message).GetFieldByName("field").(*dynamic.Message)

					// Only support OFPXMC_OPENFLOW_BASIC (0x8000)
					if set.GetFieldByName("oxm_class").(int32) != 0x8000 {
						continue
					}
					basic := set.GetFieldByName("ofb_field").(*dynamic.Message)

					switch basic.GetFieldByName("type").(int32) {
					case 6: // VLAN_ID
						f.Set(FLOW_FIELD_SET_VLAN_ID)
						f.SetVlanId = toVlanId(basic.GetFieldByName("vlan_vid").(uint32))
					default: // Unsupported
						/*
						 * For unsupported match types put them into an
						 * "Unsupported field so the table/json still
						 * outputs relatively correctly as opposed to
						 * having log messages.
						 */
						f.Set(FLOW_FIELD_UNSUPPORTED_SET_FIELD)
						f.UnsupportedSetField = appendInt32(f.UnsupportedSetField,
							basic.GetFieldByName("type").(int32))
					}
				default: // Unsupported
					/*
					 * For unsupported match types put them into an
					 * "Unsupported field so the table/json still
					 * outputs relatively correctly as opposed to
					 * having log messages.
					 */
					f.Set(FLOW_FIELD_UNSUPPORTED_ACTION)
					f.UnsupportedAction = appendInt32(f.UnsupportedAction,
						a.GetFieldByName("type").(int32))
				}
			}
		case 5: // CLEAR_ACTIONS
			// Following current CLI, just assigning empty list
			f.Set(FLOW_FIELD_CLEAR_ACTIONS)
			f.ClearActions = "[]"
		default: // Unsupported
			/*
			 * For unsupported match types put them into an
			 * "Unsupported field so the table/json still
			 * outputs relatively correctly as opposed to
			 * having log messages.
			 */
			f.Set(FLOW_FIELD_UNSUPPORTED_INSTRUCTION)
			f.UnsupportedInstruction = appendUint32(f.UnsupportedInstruction,
				inst.GetFieldByName("type").(uint32))
		}
	}
}
