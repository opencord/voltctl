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
package config

import (
	"time"
)

func (s *StackConfigSpec) GetName() string {
	return s.Name
}

func (s *StackConfigSpec) SetName(name string) {
	s.Name = name
}

func (s *StackConfigSpec) GetServer() string {
	return s.Server
}

func (s *StackConfigSpec) SetServer(server string) {
	s.Server = server
}

func (s *StackConfigSpec) GetKafka() string {
	return s.Kafka
}

func (s *StackConfigSpec) SetKafka(kafka string) {
	s.Kafka = kafka
}

func (s *StackConfigSpec) GetKvStore() string {
	return s.KvStore
}

func (s *StackConfigSpec) SetKvStore(store string) {
	s.KvStore = store
}

func (s *StackConfigSpec) GetKvStoreTimeout() time.Duration {
	return s.KvStoreConfig.Timeout
}

func (s *StackConfigSpec) SetKvStoreTimeout(t time.Duration) {
	s.KvStoreConfig.Timeout = t
}

func (s *StackConfigSpec) GetGrpcTimeout() time.Duration {
	return s.Grpc.Timeout
}

func (s *StackConfigSpec) SetGrpcTimeout(t time.Duration) {
	s.Grpc.Timeout = t
}

func (s *StackConfigSpec) GetGrpcMaxCallRecvMsgSize() string {
	return s.Grpc.MaxCallRecvMsgSize
}

func (s *StackConfigSpec) SetGrpcMaxCallRecvMsgSize(size string) {
	s.Grpc.MaxCallRecvMsgSize = size
}
