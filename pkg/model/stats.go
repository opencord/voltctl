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
package model

type OnuStats struct {
	IntfId                    *uint32 `json:"intfid,omitempty"`
	OnuId                     *uint32 `json:"onuid,omitempty"`
	PositiveDrift             *uint64 `json:"positivedrift,omitempty"`
	NegativeDrift             *uint64 `json:"negativedrift,omitempty"`
	DelimiterMissDetection    *uint64 `json:"delimitermissdetection,omitempty"`
	BipErrors                 *uint64 `json:"biperrors,omitempty"`
	BipUnits                  *uint64 `json:"bipunits,omitempty"`
	FecCorrectedSymbols       *uint64 `json:"feccorrectedsymbols,omitempty"`
	FecCodewordsCorrected     *uint64 `json:"feccodewordscorrected,omitempty"`
	FecCodewordsUncorrectable *uint64 `json:"feccodewordsuncorrectable,omitempty"`
	FecCodewords              *uint64 `json:"feccodewords,omitempty"`
	FecCorrectedUnits         *uint64 `json:"feccorrectedunits,omitempty"`
	XgemKeyErrors             *uint64 `json:"xgemkeyerrors,omitempty"`
	XgemLoss                  *uint64 `json:"xgemloss,omitempty"`
	RxPloamsError             *uint64 `json:"rxploamserror,omitempty"`
	RxPloamsNonIdle           *uint64 `json:"rxploamsnonidle,omitempty"`
	RxOmci                    *uint64 `json:"rxomci,omitempty"`
	TxOmci                    *uint64 `json:"txomci,omitempty"`
	RxOmciPacketsCrcError     *uint64 `json:"rxomcipacketscrcerror,omitempty"`
	RxBytes                   *uint64 `json:"rxbytes,omitempty"`
	RxPackets                 *uint64 `json:"rxpackets,omitempty"`
	TxBytes                   *uint64 `json:"txbytes,omitempty"`
	TxPackets                 *uint64 `json:"txpackets,omitempty"`
	BerReported               *uint64 `json:"berreported,omitempty"`
	LcdgErrors                *uint64 `json:"lcdgerrors,omitempty"`
	RdiErrors                 *uint64 `json:"rdierrors,omitempty"`
	// reported timestamp in seconds since epoch
	Timestamp *uint32 `json:"timestamp,omitempty"`
}

type OnuEthernetFrameExtendedPm struct {
	UDropEvents               *uint64 `json:"upstream_drop_events,omitempty"`
	UOctets                   *uint64 `json:"upstream_octets,omitempty"`
	UFrames                   *uint64 `json:"upstream_frames,omitempty"`
	UBroadcastFrames          *uint64 `json:"upstream_broadcast_frames,omitempty"`
	UMulticastFrames          *uint64 `json:"upstream_multicast_frames,omitempty"`
	UCrcErroredFrames         *uint64 `json:"upstream_crc_errored_frames,omitempty"`
	UUndersizeFrames          *uint64 `json:"upstream_undersize_frames,omitempty"`
	UOversizeFrames           *uint64 `json:"upstream_oversize_frames,omitempty"`
	UFrames_64Octets          *uint64 `json:"upstream_frames_64_octets,omitempty"`
	UFrames_65To_127Octets    *uint64 `json:"upstream_frames_65_to_127_octets,omitempty"`
	UFrames_128To_255Octets   *uint64 `json:"upstream_frames_128_to_255_octets,omitempty"`
	UFrames_256To_511Octets   *uint64 `json:"upstream_frames_256_to_511_octets,omitempty"`
	UFrames_512To_1023Octets  *uint64 `json:"upstream_frames_512_to_1023_octets,omitempty"`
	UFrames_1024To_1518Octets *uint64 `json:"upstream_frames_1024_to_1518_octets,omitempty"`
	DDropEvents               *uint64 `json:"downstream_drop_events,omitempty"`
	DOctets                   *uint64 `json:"downstream_octets,omitempty"`
	DFrames                   *uint64 `json:"downstream_frames,omitempty"`
	DBroadcastFrames          *uint64 `json:"downstream_broadcast_frames,omitempty"`
	DMulticastFrames          *uint64 `json:"downstream_multicast_frames,omitempty"`
	DCrcErroredFrames         *uint64 `json:"downstream_crc_errored_frames,omitempty"`
	DUndersizeFrames          *uint64 `json:"downstream_undersize_frames,omitempty"`
	DOversizeFrames           *uint64 `json:"downstream_oversize_frames,omitempty"`
	DFrames_64Octets          *uint64 `json:"downstream_frames_64_octets,omitempty"`
	DFrames_65To_127Octets    *uint64 `json:"downstream_frames_65_to_127_octets,omitempty"`
	DFrames_128To_255Octets   *uint64 `json:"downstream_frames_128_to_255_octets,omitempty"`
	DFrames_256To_511Octets   *uint64 `json:"downstream_frames_256_to_511_octets,omitempty"`
	DFrames_512To_1023Octets  *uint64 `json:"downstream_frames_512_to_1023_octets,omitempty"`
	DFrames_1024To_1518Octets *uint64 `json:"downstream_frames_1024_to_1518_octets,omitempty"`
}
