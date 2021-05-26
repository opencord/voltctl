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
