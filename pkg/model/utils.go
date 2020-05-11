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
	"errors"
	"github.com/jhump/protoreflect/dynamic"
)

func GetEnumValue(val *dynamic.Message, name string) (string, error) {
	fd := val.FindFieldDescriptorByName(name)
	if fd == nil {
		return "", errors.New("fieldDescriptor is nil for " + name)
	}

	enumType := fd.GetEnumType()
	if enumType == nil {
		return "", errors.New("enumType is nil for " + name)
	}

	field, ok := val.GetFieldByName(name).(int32)
	if !ok {
		return "", errors.New("Enum integer value not found for " + name)
	}

	eValue := enumType.FindValueByNumber(field)
	if eValue == nil {
		return "", errors.New("Value not found for " + name)
	}

	return eValue.GetName(), nil
}

func SetEnumValue(msg *dynamic.Message, name string, value string) error {
	fd := msg.FindFieldDescriptorByName(name)
	if fd == nil {
		return errors.New("fieldDescriptor is nil for " + name)
	}

	enumType := fd.GetEnumType()
	if enumType == nil {
		return errors.New("enumType is nil for " + name)
	}

	eValue := enumType.FindValueByName(value)
	if eValue == nil {
		return errors.New("Value not found for " + name)
	}

	msg.SetFieldByName(name, eValue.GetNumber())
	return nil
}

func GetEnumString(msg *dynamic.Message, name string, value int32) (string, error) {
	fd := msg.FindFieldDescriptorByName(name)
	if fd == nil {
		return "", errors.New("fieldDescriptor is nil for " + name)
	}

	enumType := fd.GetEnumType()
	if enumType == nil {
		return "", errors.New("enumType is nil for " + name)
	}

	eValue := enumType.FindValueByNumber(value)
	if eValue == nil {
		return "", errors.New("Value not found for " + name)
	}
	return eValue.GetName(), nil
}
