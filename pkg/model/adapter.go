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
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jhump/protoreflect/dynamic"
	"time"
)

type Adapter struct {
	Id                     string
	Vendor                 string
	Version                string
	LogLevel               string
	LastCommunication      string
	SinceLastCommunication string
	CurrentReplica         int32
	TotalReplicas          int32
	Endpoint               string
	Type                   string
}

func (adapter *Adapter) PopulateFrom(val *dynamic.Message) {

	adapter.Id = val.GetFieldByName("id").(string)
	adapter.Vendor = val.GetFieldByName("vendor").(string)
	adapter.Version = val.GetFieldByName("version").(string)

	// TODO use TryGetFieldByName as these fields are not in V1/V2


	if value, err := val.TryGetFieldByName("currentReplica"); err != nil {
		adapter.CurrentReplica = 0
	} else {
		adapter.CurrentReplica = value.(int32)
	}
	if value, err := val.TryGetFieldByName("totalReplicas"); err != nil {
		adapter.TotalReplicas = 0
	} else {
		adapter.TotalReplicas = value.(int32)
	}
	if value, err := val.TryGetFieldByName("endpoint"); err != nil {
		adapter.Endpoint = "UNKNOWN"
	} else {
		adapter.Endpoint = value.(string)
	}
	if value, err := val.TryGetFieldByName("type"); err != nil {
		adapter.Type = "UNKNOWN"
	} else {
		adapter.Type = value.(string)
	}

	if lastCommunication, err := val.TryGetFieldByName("last_communication"); err != nil {
		adapter.LastCommunication = "UNKNOWN"
		adapter.SinceLastCommunication = "UNKNOWN"
	} else {
		if lastCommunication, err := ptypes.Timestamp(lastCommunication.(*timestamp.Timestamp)); err != nil {
			adapter.LastCommunication = "NEVER"
			adapter.SinceLastCommunication = "NEVER"
		} else {
			adapter.LastCommunication = lastCommunication.Truncate(time.Second).Format(time.RFC3339)
			adapter.SinceLastCommunication = time.Since(lastCommunication).Truncate(time.Second).String()
		}
	}

}
