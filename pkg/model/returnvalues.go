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
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

type ReturnValueRow struct {
	Name   string      `json:"name"`
	Result interface{} `json:"result"`
}

type ReturnValues struct {
	Set         uint32 `json:"set"`
	Unsupported uint32 `json:"unsupported"`
	Error       uint32 `json:"error"`
	Distance    uint32 `json:"distance"`
}

func (r *ReturnValues) PopulateFrom(val *dynamic.Message) {
	r.Set = val.GetFieldByName("Set").(uint32)
	r.Unsupported = val.GetFieldByName("Unsupported").(uint32)
	r.Error = val.GetFieldByName("Error").(uint32)
	r.Distance = val.GetFieldByName("Distance").(uint32)
}

// Given a list of allowed enum values, check each one of the values against the
// bitmaps, and fill out an array of result rows as follows:
//    "Error", if the enum is set in the Error bitmap
//    "Unsupported", if the enum is set in the Unsupported bitmap
//    An interface containing the value, if the enum is set in the Set bitmap

func (r *ReturnValues) GetKeyValuePairs(enumValues []*desc.EnumValueDescriptor) []ReturnValueRow {
	var rows []ReturnValueRow

	for _, v := range enumValues {
		num := uint32(v.GetNumber())
		if num == 0 {
			// EMPTY is not a real value
			continue
		}
		name := v.GetName()
		if (r.Error & num) != 0 {
			row := ReturnValueRow{Name: name, Result: "Error"}
			rows = append(rows, row)
		}
		if (r.Unsupported & num) != 0 {
			row := ReturnValueRow{Name: name, Result: "Unsupported"}
			rows = append(rows, row)
		}
		if (r.Set & num) != 0 {
			switch name {
			case "DISTANCE":
				row := ReturnValueRow{Name: name, Result: r.Distance}
				rows = append(rows, row)
			default:
				row := ReturnValueRow{Name: name, Result: "Unimplemented-in-voltctl"}
				rows = append(rows, row)
			}
		}
	}

	return rows
}
