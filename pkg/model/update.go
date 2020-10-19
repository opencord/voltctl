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
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/opencord/voltha-protos/v4/go/common"
	"github.com/opencord/voltha-protos/v4/go/voltha"
)

/*
 * This is a partial list of OF match/action values. This list will be
 * expanded as new fields are needed within VOLTHA
 *
 * Strings are used in the output structure so that on output the table
 * can be "sparsely" populated with "empty" cells as opposed to 0 (zeros)
 * all over the place.
 */
type Update struct {
	Timestamp   timestamp.Timestamp `json:"timestamp"`
	Operation   string              `json:"operation"`
	OperationID string              `json:"operationId"`
	RequestedBy string              `json:"requestedBy"`
	StateChange string              `json:"stateChange"`
	Status      string              `json:"status"`
	Description string              `json:"description"`
}

func (f *Update) PopulateFromProto(update *voltha.DeviceUpdate) {
	f.Timestamp = *update.Timestamp
	f.Operation = update.Operation
	f.OperationID = update.OperationId
	f.RequestedBy = update.RequestedBy
	if update.StateChange.Current.AdminState != update.StateChange.Previous.AdminState {
		f.StateChange = fmt.Sprintf("%v->%v", update.StateChange.Previous.AdminState,
			update.StateChange.Current.AdminState)
	}
	f.Status = common.OperationResp_OperationReturnCode_name[int32(update.Status.Code)]
	f.Description = update.Description
}
