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

func GetEnumValue(val *dynamic.Message, name string) string {
	return val.FindFieldDescriptorByName(name).GetEnumType().
		FindValueByNumber(val.GetFieldByName(name).(int32)).GetName()
}

func SetEnumValue(msg *dynamic.Message, name string, value string) {
	eValue := msg.FindFieldDescriptorByName(name).GetEnumType().FindValueByName(value)
	msg.SetFieldByName(name, eValue.GetNumber())
}

func GetEnumString(msg *dynamic.Message, name string, value int32) string {
	eValue := msg.FindFieldDescriptorByName(name).GetEnumType().FindValueByNumber(value)
	if eValue == nil {
		panic("eValue is nil")
	}
	return eValue.GetName()
}
