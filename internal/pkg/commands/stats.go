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
	"fmt"
	"github.com/opencord/voltctl/pkg/model"
	"github.com/opencord/voltha-protos/v5/go/extension"
	"strings"
)

type tagBuilder struct {
	firstField bool
	tag        strings.Builder
}

func NewTagBuilder() tagBuilder {
	return tagBuilder{
		firstField: true,
		tag:        strings.Builder{},
	}
}
func (t *tagBuilder) buildOutputString() string {
	return t.tag.String()
}
func (t *tagBuilder) addFieldInFormat(name string) {
	if !t.firstField {
		t.tag.WriteString("\n")
	}
	t.firstField = false
	str := fmt.Sprintf("%s:    ", name)
	t.tag.WriteString(str)
	t.tag.WriteString("{{.")
	t.tag.WriteString(name)
	t.tag.WriteString("}}")
}

/*
 * Construct a template format string based on the fields required by the
 * results.
 */
func buildOnuStatsOutputFormat(counters *extension.GetOnuCountersResponse) (model.OnuStats, string) {
	onuStats := model.OnuStats{}
	tagBuilder := NewTagBuilder()
	if counters.IsIntfId != nil {
		intfId := counters.GetIntfId()
		onuStats.IntfId = &intfId
		tagBuilder.addFieldInFormat("IntfId")
	}
	if counters.IsOnuId != nil {
		onuId := counters.GetOnuId()
		onuStats.OnuId = &onuId
		tagBuilder.addFieldInFormat("OnuId")
	}
	if counters.IsPositiveDrift != nil {
		positiveDrift := counters.GetPositiveDrift()
		onuStats.PositiveDrift = &positiveDrift
		tagBuilder.addFieldInFormat("PositiveDrift")
	}
	if counters.IsNegativeDrift != nil {
		negativeDrift := counters.GetNegativeDrift()
		onuStats.NegativeDrift = &negativeDrift
		tagBuilder.addFieldInFormat("NegativeDrift")
	}
	if counters.IsDelimiterMissDetection != nil {
		delimiterMissDet := counters.GetDelimiterMissDetection()
		onuStats.DelimiterMissDetection = &delimiterMissDet
		tagBuilder.addFieldInFormat("DelimiterMissDetection")
	}
	if counters.IsBipErrors != nil {
		bipErrors := counters.GetBipErrors()
		onuStats.BipErrors = &bipErrors
		tagBuilder.addFieldInFormat("BipErrors")
	}
	if counters.IsBipUnits != nil {
		bipUnits := counters.GetBipUnits()
		onuStats.BipUnits = &bipUnits
		tagBuilder.addFieldInFormat("BipUnits")
	}
	if counters.IsFecCorrectedSymbols != nil {
		fecCorrectedSymbols := counters.GetFecCorrectedSymbols()
		onuStats.FecCorrectedSymbols = &fecCorrectedSymbols
		tagBuilder.addFieldInFormat("FecCorrectedSymbols")
	}
	if counters.IsFecCodewordsCorrected != nil {
		fecCodewordsCorrected := counters.GetFecCodewordsCorrected()
		onuStats.FecCodewordsCorrected = &fecCodewordsCorrected
		tagBuilder.addFieldInFormat("FecCodewordsCorrected")
	}
	if counters.IsFecCodewordsUncorrectable != nil {
		fecCodewordsUncorrectable := counters.GetFecCodewordsUncorrectable()
		onuStats.FecCodewordsUncorrectable = &fecCodewordsUncorrectable
		tagBuilder.addFieldInFormat("FecCodewordsUncorrectable")
	}
	if counters.IsFecCodewords != nil {
		fecCodewords := counters.GetFecCodewords()
		onuStats.FecCodewords = &fecCodewords
		tagBuilder.addFieldInFormat("FecCodewords")
	}
	if counters.IsFecCorrectedUnits != nil {
		fecCorrectedUnits := counters.GetFecCorrectedUnits()
		onuStats.FecCorrectedUnits = &fecCorrectedUnits
		tagBuilder.addFieldInFormat("FecCorrectedUnits")
	}
	if counters.IsXgemKeyErrors != nil {
		xgemKeyErrors := counters.GetXgemKeyErrors()
		onuStats.XgemKeyErrors = &xgemKeyErrors
		tagBuilder.addFieldInFormat("XgemKeyErrors")
	}
	if counters.IsXgemLoss != nil {
		xgemLoss := counters.GetXgemLoss()
		onuStats.XgemLoss = &xgemLoss
		tagBuilder.addFieldInFormat("XgemLoss")
	}
	if counters.IsRxPloamsError != nil {
		rxPloamsError := counters.GetRxPloamsError()
		onuStats.RxPloamsError = &rxPloamsError
		tagBuilder.addFieldInFormat("RxPloamsError")
	}
	if counters.IsRxPloamsNonIdle != nil {
		rxPloamsNonIdle := counters.GetRxPloamsNonIdle()
		onuStats.RxPloamsNonIdle = &rxPloamsNonIdle
		tagBuilder.addFieldInFormat("RxPloamsNonIdle")
	}
	if counters.IsRxOmci != nil {
		rxOmci := counters.GetRxOmci()
		onuStats.RxOmci = &rxOmci
		tagBuilder.addFieldInFormat("RxOmci")
	}
	if counters.IsTxOmci != nil {
		txOmci := counters.GetTxOmci()
		onuStats.TxOmci = &txOmci
		tagBuilder.addFieldInFormat("TxOmci")
	}
	if counters.IsRxOmciPacketsCrcError != nil {
		rxOmciPacketsCrcError := counters.GetRxOmciPacketsCrcError()
		onuStats.RxOmciPacketsCrcError = &rxOmciPacketsCrcError
		tagBuilder.addFieldInFormat("RxOmciPacketsCrcError")
	}
	if counters.IsRxBytes != nil {
		rxBytes := counters.GetRxBytes()
		onuStats.RxBytes = &rxBytes
		tagBuilder.addFieldInFormat("RxBytes")
	}
	if counters.IsRxPackets != nil {
		rxPackets := counters.GetRxPackets()
		onuStats.RxPackets = &rxPackets
		tagBuilder.addFieldInFormat("RxPackets")
	}
	if counters.IsTxBytes != nil {
		txBytes := counters.GetTxBytes()
		onuStats.TxBytes = &txBytes
		tagBuilder.addFieldInFormat("TxBytes")
	}
	if counters.IsTxPackets != nil {
		txPackets := counters.GetTxPackets()
		onuStats.TxPackets = &txPackets
		tagBuilder.addFieldInFormat("TxPackets")
	}
	if counters.IsBerReported != nil {
		berReported := counters.GetBerReported()
		onuStats.BerReported = &berReported
		tagBuilder.addFieldInFormat("BerReported")
	}
	if counters.IsLcdgErrors != nil {
		lcdgErrors := counters.GetLcdgErrors()
		onuStats.LcdgErrors = &lcdgErrors
		tagBuilder.addFieldInFormat("LcdgErrors")
	}
	if counters.IsRdiErrors != nil {
		rdiErrors := counters.GetRdiErrors()
		onuStats.RdiErrors = &rdiErrors
		tagBuilder.addFieldInFormat("RdiErrors")
	}
	if counters.IsTimestamp != nil {
		timestamp := counters.GetTimestamp()
		onuStats.Timestamp = &timestamp
		tagBuilder.addFieldInFormat("Timestamp")
	}
	return onuStats, tagBuilder.buildOutputString()
}

/*
 * Construct a template format string based on the fields required by the
 * results.
 */
func buildOnuEthernetFrameExtendedPmOutputFormat(counters *extension.GetOmciEthernetFrameExtendedPmResponse) model.OnuEthernetFrameExtendedPm {
	onuStats := model.OnuEthernetFrameExtendedPm{}

	dropEvents := counters.Upstream.GetDropEvents()
	onuStats.UDropEvents = &dropEvents

	octets := counters.Upstream.GetOctets()
	onuStats.UOctets = &octets

	frames := counters.Upstream.GetFrames()
	onuStats.UFrames = &frames

	broadcastFrames := counters.Upstream.GetBroadcastFrames()
	onuStats.UBroadcastFrames = &broadcastFrames

	multicastFrames := counters.Upstream.GetMulticastFrames()
	onuStats.UMulticastFrames = &multicastFrames

	crcErroredFrames := counters.Upstream.GetCrcErroredFrames()
	onuStats.UCrcErroredFrames = &crcErroredFrames

	undersizeFrames := counters.Upstream.GetUndersizeFrames()
	onuStats.UUndersizeFrames = &undersizeFrames

	oversizeFrames := counters.Upstream.GetOversizeFrames()
	onuStats.UOversizeFrames = &oversizeFrames

	frames_64Octets := counters.Upstream.GetFrames_64Octets()
	onuStats.UFrames_64Octets = &frames_64Octets

	frames_65To_127Octets := counters.Upstream.GetFrames_65To_127Octets()
	onuStats.UFrames_65To_127Octets = &frames_65To_127Octets

	frames_128To_255Octets := counters.Upstream.GetFrames_128To_255Octets()
	onuStats.UFrames_128To_255Octets = &frames_128To_255Octets

	frames_256To_511Octets := counters.Upstream.GetFrames_256To_511Octets()
	onuStats.UFrames_256To_511Octets = &frames_256To_511Octets

	frames_512To_1023Octets := counters.Upstream.GetFrames_512To_1023Octets()
	onuStats.UFrames_512To_1023Octets = &frames_512To_1023Octets

	frames_1024To_1518Octets := counters.Upstream.GetFrames_1024To_1518Octets()
	onuStats.UFrames_1024To_1518Octets = &frames_1024To_1518Octets

	dDropEvents := counters.Downstream.GetDropEvents()
	onuStats.DDropEvents = &dDropEvents

	dOctets := counters.Downstream.GetOctets()
	onuStats.DOctets = &dOctets

	dFrames := counters.Downstream.GetFrames()
	onuStats.DFrames = &dFrames

	dBroadcastFrames := counters.Downstream.GetBroadcastFrames()
	onuStats.DBroadcastFrames = &dBroadcastFrames

	dMulticastFrames := counters.Downstream.GetMulticastFrames()
	onuStats.DMulticastFrames = &dMulticastFrames

	dCrcErroredFrames := counters.Downstream.GetCrcErroredFrames()
	onuStats.DCrcErroredFrames = &dCrcErroredFrames

	dUndersizeFrames := counters.Downstream.GetUndersizeFrames()
	onuStats.DUndersizeFrames = &dUndersizeFrames

	dOversizeFrames := counters.Downstream.GetOversizeFrames()
	onuStats.DOversizeFrames = &dOversizeFrames

	dFrames_64Octets := counters.Downstream.GetFrames_64Octets()
	onuStats.DFrames_64Octets = &dFrames_64Octets

	dFrames_65To_127Octets := counters.Downstream.GetFrames_65To_127Octets()
	onuStats.DFrames_65To_127Octets = &dFrames_65To_127Octets

	dFrames_128To_255Octets := counters.Downstream.GetFrames_128To_255Octets()
	onuStats.DFrames_128To_255Octets = &dFrames_128To_255Octets

	dFrames_256To_511Octets := counters.Downstream.GetFrames_256To_511Octets()
	onuStats.DFrames_256To_511Octets = &dFrames_256To_511Octets

	dFrames_512To_1023Octets := counters.Downstream.GetFrames_512To_1023Octets()
	onuStats.DFrames_512To_1023Octets = &dFrames_512To_1023Octets

	dFrames_1024To_1518Octets := counters.Downstream.GetFrames_1024To_1518Octets()
	onuStats.DFrames_1024To_1518Octets = &dFrames_1024To_1518Octets
	return onuStats
}
