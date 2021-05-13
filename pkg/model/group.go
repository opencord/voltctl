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
	"github.com/opencord/voltha-protos/v4/go/openflow_13"
)

type GroupFieldFlag uint64

const (
	// Define bit flags for group fields to determine what is set and
	// what is not
	GROUP_FIELD_UNSUPPORTED_ACTION GroupFieldFlag = 1 << iota
	GROUP_FIELD_UNSUPPORTED_SET_FIELD
	GROUP_FIELD_GROUP_ID
	GROUP_FIELD_GROUP_TYPE
	GROUP_FIELD_DURATION_SEC
	GROUP_FIELD_DURATION_NSEC
	GROUP_FIELD_REF_COUNT
	GROUP_FIELD_PACKET_COUNT
	GROUP_FIELD_BYTE_COUNT
	GROUP_FIELD_BUCKETS
	GROUP_FIELD_WEIGHT
	GROUP_FIELD_WATCH_GROUP
	GROUP_FIELD_WATCH_PORT
	GROUP_FIELD_OUTPUT
	GROUP_FIELD_PUSH_VLAN_ID
	GROUP_FIELD_POP_VLAN
	GROUP_FIELD_SET_VLAN_ID
	GROUP_FIELD_BUCKET_STATS
	GROUP_FIELD_BUCKET_STATS_PACKET_COUNT
	GROUP_FIELD_BUCKET_STATS_BYTE_COUNT

	GROUP_FIELD_HEADER = GROUP_FIELD_GROUP_ID | GROUP_FIELD_GROUP_TYPE

	GROUP_FIELD_STATS = GROUP_FIELD_DURATION_SEC | GROUP_FIELD_DURATION_NSEC |
		GROUP_FIELD_REF_COUNT | GROUP_FIELD_PACKET_COUNT | GROUP_FIELD_BYTE_COUNT

)

var (
	// Provide an array of all flags that can be used for iteration
	AllGroupFieldFlags = []GroupFieldFlag{
		GROUP_FIELD_UNSUPPORTED_ACTION,
		GROUP_FIELD_UNSUPPORTED_SET_FIELD,
		GROUP_FIELD_GROUP_ID,
		GROUP_FIELD_GROUP_TYPE,
		GROUP_FIELD_DURATION_SEC,
		GROUP_FIELD_DURATION_NSEC,
		GROUP_FIELD_REF_COUNT,
		GROUP_FIELD_PACKET_COUNT,
		GROUP_FIELD_BYTE_COUNT,
		GROUP_FIELD_BUCKETS,
		GROUP_FIELD_WEIGHT,
		GROUP_FIELD_WATCH_GROUP,
		GROUP_FIELD_WATCH_PORT,
		GROUP_FIELD_OUTPUT,
		GROUP_FIELD_PUSH_VLAN_ID,
		GROUP_FIELD_POP_VLAN,
		GROUP_FIELD_SET_VLAN_ID,
		GROUP_FIELD_BUCKET_STATS,
		GROUP_FIELD_BUCKET_STATS_PACKET_COUNT,
		GROUP_FIELD_BUCKET_STATS_BYTE_COUNT,
	}
)

func (f *GroupFieldFlag) Count() int {
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
func (f *GroupFieldFlag) IsSet(flag GroupFieldFlag) bool {
	return *f&flag > 0
}

func (f *GroupFieldFlag) Set(flag GroupFieldFlag) {
	*f |= flag
}

func (f *GroupFieldFlag) Clear(flag GroupFieldFlag) {
	var mask = ^(flag)
	*f &= mask
}

func (f *GroupFieldFlag) Reset() {
	*f = 0
}

func (f GroupFieldFlag) String() string {
	switch f {
	case GROUP_FIELD_UNSUPPORTED_ACTION:
		return "UnsupportedAction"
	case GROUP_FIELD_UNSUPPORTED_SET_FIELD:
		return "UnsupportedSetField"
	case GROUP_FIELD_GROUP_ID:
		return "GroupId"
	case GROUP_FIELD_GROUP_TYPE:
		return "GroupType"
	case GROUP_FIELD_DURATION_SEC:
		return "DurationSec"
	case GROUP_FIELD_DURATION_NSEC:
		return "DurationNsec"
	case GROUP_FIELD_REF_COUNT:
		return "RefCount"
	case GROUP_FIELD_PACKET_COUNT:
		return "PacketCount"
	case GROUP_FIELD_BYTE_COUNT:
		return "ByteCount"
	case GROUP_FIELD_BUCKETS:
		return "Buckets"
	case GROUP_FIELD_WEIGHT:
		return "Weight"
	case GROUP_FIELD_WATCH_GROUP:
		return "WatchGroup"
	case GROUP_FIELD_WATCH_PORT:
		return "WatchPort"
	case GROUP_FIELD_OUTPUT:
		return "Output"
	case GROUP_FIELD_PUSH_VLAN_ID:
		return "PushVlanId"
	case GROUP_FIELD_POP_VLAN:
		return "PopVlan"
	case GROUP_FIELD_SET_VLAN_ID:
		return "SetVlanId"
	case GROUP_FIELD_BUCKET_STATS:
		return "BucketStats"
	case GROUP_FIELD_BUCKET_STATS_PACKET_COUNT:
		return "PacketCount"
	case GROUP_FIELD_BUCKET_STATS_BYTE_COUNT:
		return "ByteCount"
	default:
		return "UnknownFieldFlag"
	}
}

func (f GroupFieldFlag) GetFormatString() string {
	switch f {
	case GROUP_FIELD_UNSUPPORTED_ACTION:
		return "UnsupportedAction"
	case GROUP_FIELD_UNSUPPORTED_SET_FIELD:
		return "UnsupportedSetField"
	case GROUP_FIELD_GROUP_ID:
		return "GroupId"
	case GROUP_FIELD_GROUP_TYPE:
		return "GroupType"
	case GROUP_FIELD_DURATION_SEC:
		return "DurationSec"
	case GROUP_FIELD_DURATION_NSEC:
		return "DurationNsec"
	case GROUP_FIELD_REF_COUNT:
		return "RefCount"
	case GROUP_FIELD_PACKET_COUNT:
		return "PacketCount"
	case GROUP_FIELD_BYTE_COUNT:
		return "ByteCount"
	case GROUP_FIELD_BUCKETS:
		return "Buckets"
	case GROUP_FIELD_WEIGHT:
		return "Buckets[0].Weight"
	case GROUP_FIELD_WATCH_GROUP:
		return "Buckets[0].WatchGroup"
	case GROUP_FIELD_WATCH_PORT:
		return "Buckets[0].WatchPort"
	case GROUP_FIELD_OUTPUT:
		return "Buckets[0].Output"
	case GROUP_FIELD_PUSH_VLAN_ID:
		return "Buckets[0].PushVlanId"
	case GROUP_FIELD_POP_VLAN:
		return "Buckets[0].PopVlan"
	case GROUP_FIELD_SET_VLAN_ID:
		return "Buckets[0].SetVlanId"
	case GROUP_FIELD_BUCKET_STATS:
		return "BucketStats"
	case GROUP_FIELD_BUCKET_STATS_PACKET_COUNT:
		return "BucketStats[0].PacketCount"
	case GROUP_FIELD_BUCKET_STATS_BYTE_COUNT:
		return "BucketStats[0].ByteCount"
	default:
		return "UnknownFieldFlag"
	}
}

type Bucket struct {
	Weight              uint32 `json:"weight"`
	WatchPort           uint32 `json:"watchport"`
	WatchGroup          uint32 `json:"watchgroup"`
	Output              string `json:"output,omitempty"`
	SetVlanId           string `json:"setvlanid,omitempty"`
	PopVlan             string `json:"popvlan,omitempty"`
	PushVlanId          string `json:"pushvlanid,omitempty"`
	UnsupportedAction   string `json:"unsupportedaction,omitempty"`
	UnsupportedSetField string `json:"unsupportedsetfield,omitempty"`
}

type BucketCounter struct {
	PacketCount uint64 `json:"packetcount"`
	ByteCount   uint64 `json:"bytecount"`
}
type Group struct {
	GroupId   uint32   `json:"groupid"`
	GroupType string   `json:"grouptype"`
	Buckets   []Bucket `json:"buckets,omitempty"`

	RefCount     uint32          `json:"refcount"`
	PacketCount  uint64          `json:"packetcount"`
	ByteCount    uint64          `json:"bytecount"`
	DurationSec  uint32          `json:"durationsec"`
	DurationNsec uint32          `json:"durationnsec"`
	BucketStats  []BucketCounter `json:"bucketstats,omitempty"`

	populated GroupFieldFlag
}

func (g *Group) Count() int {
	return g.populated.Count()
}

func (g *Group) IsSet(flag GroupFieldFlag) bool {
	return g.populated.IsSet(flag)
}

func (g *Group) Set(flag GroupFieldFlag) {
	g.populated.Set(flag)
}

func (g *Group) Clear(flag GroupFieldFlag) {
	g.populated.Clear(flag)
}

func (g *Group) Reset() {
	g.populated.Reset()
}

func (g *Group) Populated() GroupFieldFlag {
	return g.populated
}

func (g *Group) PopulateGroupFormatFromProto(group *openflow_13.OfpGroupEntry) {

	g.Reset()

	// Fill desc first
	g.GroupId = group.Desc.GroupId
	g.GroupType = group.Desc.Type.String()
	g.Set(GROUP_FIELD_HEADER)
	g.Buckets = make([]Bucket, len(group.Desc.Buckets))
	if len(group.Desc.Buckets) > 0 {
		g.Set(GROUP_FIELD_BUCKETS)
		g.Set(GROUP_FIELD_WEIGHT)
		g.Set(GROUP_FIELD_WATCH_PORT)
		g.Set(GROUP_FIELD_WATCH_GROUP)
	}
	for _, bucket := range group.Desc.Buckets {
		var bckt Bucket
		bckt.WatchGroup = bucket.WatchGroup
		bckt.WatchPort = bucket.WatchPort
		bckt.Weight = bucket.Weight
		for _, action := range bucket.Actions {
			switch action.Type {
			case 0: // OUTPUT
				g.Set(GROUP_FIELD_OUTPUT)
				output := action.GetOutput()
				out := output.Port
				switch out & 0x7fffffff {
				case 0:
					bckt.Output = "INVALID"
				case 0x7ffffff8:
					bckt.Output = "IN_PORT"
				case 0x7ffffff9:
					bckt.Output = "TABLE"
				case 0x7ffffffa:
					bckt.Output = "NORMAL"
				case 0x7ffffffb:
					bckt.Output = "FLOOD"
				case 0x7ffffffc:
					bckt.Output = "ALL"
				case 0x7ffffffd:
					bckt.Output = "CONTROLLER"
				case 0x7ffffffe:
					bckt.Output = "LOCAL"
				case 0x7fffffff:
					bckt.Output = "ANY"
				default:
					bckt.Output = fmt.Sprintf("%d", output.Port)
				}
			case 17: // PUSH_VLAN
				g.Set(GROUP_FIELD_PUSH_VLAN_ID)
				push := action.GetPush()
				bckt.PushVlanId = fmt.Sprintf("0x%x", push.Ethertype)
			case 18: // POP_VLAN
				g.Set(GROUP_FIELD_POP_VLAN)
				bckt.PopVlan = "yes"
			case 25: // SET_FIELD
				set := action.GetSetField().Field

				// Only support OFPXMC_OPENFLOW_BASIC (0x8000)
				if set.OxmClass != 0x8000 {
					continue
				}
				basic := set.GetOfbField()

				switch basic.Type {
				case 6: // VLAN_ID
					g.Set(GROUP_FIELD_SET_VLAN_ID)
					bckt.SetVlanId = toVlanId(basic.GetVlanVid())
				default: // Unsupported
					/*
					 * For unsupported match types put them into an
					 * "Unsupported field so the table/json still
					 * outputs relatively correctly as opposed to
					 * having log messages.
					 */
					g.Set(GROUP_FIELD_UNSUPPORTED_SET_FIELD)
					bckt.UnsupportedSetField = appendInt32(bckt.UnsupportedSetField,
						int32(basic.Type))
				}
			default: // Unsupported
				/*
				 * For unsupported match types put them into an
				 * "Unsupported field so the table/json still
				 * outputs relatively correctly as opposed to
				 * having log messages.
				 */
				g.Set(GROUP_FIELD_UNSUPPORTED_ACTION)
				bckt.UnsupportedAction = appendInt32(bckt.UnsupportedAction,
					int32(action.Type))
			}
		}
		g.Buckets = append(g.Buckets, bckt)
	}
	g.RefCount = group.Stats.RefCount
	g.DurationSec = group.Stats.DurationSec
	g.DurationNsec = group.Stats.DurationNsec
	g.PacketCount = group.Stats.PacketCount
	g.ByteCount = group.Stats.ByteCount
	g.Set(GROUP_FIELD_STATS)
	g.BucketStats = make([]BucketCounter, len(group.Stats.BucketStats))
	if len(group.Stats.BucketStats) > 0 {
		g.Set(GROUP_FIELD_BUCKET_STATS)
		g.Set(GROUP_FIELD_BUCKET_STATS_BYTE_COUNT)
		g.Set(GROUP_FIELD_BUCKET_STATS_PACKET_COUNT)
	}
	for _, bucketStats := range group.Stats.BucketStats {
		var bcktStats BucketCounter
		bcktStats.ByteCount = bucketStats.ByteCount
		bcktStats.PacketCount = bucketStats.PacketCount
		g.BucketStats = append(g.BucketStats, bcktStats)
	}

}
