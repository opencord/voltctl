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
	"strings"
	"github.com/opencord/voltctl/pkg/model"
	"github.com/opencord/voltha-protos/v4/go/extension"
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
		t.tag.WriteString("\t")
	}
	t.firstField = false
	t.tag.WriteString("{{.")
	t.tag.WriteString(name)
	t.tag.WriteString("}}")
}
func (t *tagBuilder) addTableInFormat() {
	t.tag.WriteString("table")
}

/*
 * Construct a template format string based on the fields required by the
 * results.
 */
func buildOnuStatsOutputFormat(counters *extension.GetOnuCountersResponse) (model.OnuStats, string) {
	onuStats := model.OnuStats{}
	tagBuilder := NewTagBuilder()
	tagBuilder.addTableInFormat()
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
