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

type DeviceGroup struct {
	Id             string   `json:"id"`
	LogicalDevices []string `json:"logicaldevices"`
	Devices        []string `json:"devices"`
}

func getId(val *dynamic.Message) string {
	return val.GetFieldByName("id").(string)
}

func (d *DeviceGroup) PopulateFrom(val *dynamic.Message) {
	d.Id = val.GetFieldByName("id").(string)
	logicaldevices := val.GetFieldByName("logical_devices").([]interface{})
	d.LogicalDevices = make([]string, len(logicaldevices))
	for i, logicaldevice := range logicaldevices {
		d.LogicalDevices[i] = getId(logicaldevice.(*dynamic.Message))
	}
	devices := val.GetFieldByName("logical_devices").([]interface{})
	d.Devices = make([]string, len(devices))
	for i, device := range devices {
		d.Devices[i] = getId(device.(*dynamic.Message))
	}
}
