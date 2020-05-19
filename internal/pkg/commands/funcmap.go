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
package commands

import (
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/proto"
	descpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type DescriptorNotFoundError struct {
	Version string
}

func (e *DescriptorNotFoundError) Error() string {
	return fmt.Sprintf("Protocol buffer descriptor for API version '%s' not found", e.Version)
}

type UnableToParseDescriptorErrror struct {
	err     error
	Version string
}

func (e *UnableToParseDescriptorErrror) Error() string {
	return fmt.Sprintf("Unable to parse protocal buffer descriptor for version '%s': %s", e.Version, e.err)
}

var descriptorMap = map[string][]byte{
	"v1": V1Descriptor,
	"v2": V2Descriptor,
	"v3": V3Descriptor,
}

// Get the descriptor source using the current ApiVersion setting
func GetDescriptorSource() (grpcurl.DescriptorSource, error) {
	version := GlobalConfig.ApiVersion
	filename, ok := descriptorMap[version]
	if !ok {
		return nil, &DescriptorNotFoundError{version}
	}
	var fds descpb.FileDescriptorSet
	err := proto.Unmarshal(filename, &fds)
	if err != nil {
		return nil, &UnableToParseDescriptorErrror{err, version}
	}
	desc, err := grpcurl.DescriptorSourceFromFileDescriptorSet(&fds)
	if err != nil {
		return nil, err
	}

	return desc, nil
}
