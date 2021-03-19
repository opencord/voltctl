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

func NewDefaultConfig() *GlobalConfigSpec {
	return &GlobalConfigSpec{
		ApiVersion: "v2",
		Server:     "localhost:55555",
		Kafka:      "localhost:9092",
		KvStore:    "localhost:2379",
		Tls: TlsConfigSpec{
			UseTls: false,
			CACert: "",
			Cert:   "",
			Key:    "",
			Verify: false,
		},
		Grpc: GrpcConfigSpec{
			ConnectTimeout:     5 * time.Second,
			Timeout:            5 * time.Minute,
			MaxCallRecvMsgSize: "4MB",
		},
		KvStoreConfig: KvStoreConfigSpec{
			Timeout: 5 * time.Second,
		},
		K8sConfig: "",
	}
}
