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

type StackConfigAPI interface {
	GetName() string
	SetName(string)
	GetServer() string
	SetServer(string)
	GetKafka() string
	SetKafka(string)
	GetKvStore() string
	SetKvStore(string)
	GetKvStoreTimeout() time.Duration
	SetKvStoreTimeout(time.Duration)
	GetGrpcTimeout() time.Duration
	SetGrpcTimeout(time.Duration)
	GetGrpcMaxCallRecvMsgSize() string
	SetGrpcMaxCallRecvMsgSize(string)
}

type GlobalConfigAPI interface {
	Write(string) error
	GetStacks() []StackConfigAPI
	AddStack(string)
	DeleteStack(string) error
	GetStackByName(string) StackConfigAPI
	GetCurrentStack() string
	SetCurrentStack(string) error
	GetName() string
	SetName(string)
	GetServer() string
	SetServer(string)
	GetKafka() string
	SetKafka(string)
	GetKvStore() string
	SetKvStore(string)
	GetKvStoreTimeout() time.Duration
	SetKvStoreTimeout(time.Duration)
	GetGrpcTimeout() time.Duration
	SetGrpcTimeout(time.Duration)
	GetGrpcMaxCallRecvMsgSize() string
	SetGrpcMaxCallRecvMsgSize(string)
}
