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
	"time"

	"github.com/opencord/voltha-protos/v4/go/voltha"
)

type AdapterInstance struct {
	Id                string    `json:"id"`
	Vendor            string    `json:"vendor"`
	Type              string    `json:"type"`
	Version           string    `json:"version"`
	Endpoint          string    `json:"endpoint"`
	CurrentReplica    int32     `json:"currentreplica"`
	TotalReplicas     int32     `json:"totalreplicas"`
	LastCommunication time.Time `json:"lastcommunication"`
}

func (a *AdapterInstance) PopulateFrom(val *voltha.Adapter) {
	a.Id = val.Id
	a.Vendor = val.Vendor
	a.Type = val.Type
	a.Version = val.Version
	a.Endpoint = val.Endpoint
	a.CurrentReplica = val.CurrentReplica
	a.TotalReplicas = val.TotalReplicas
	a.LastCommunication = time.Unix(val.LastCommunication, 0)
}
