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
	"strconv"

	configv1 "github.com/opencord/voltctl/internal/pkg/apis/config/v1"
)

func FromConfigV1(v1 *configv1.GlobalConfigSpec) *GlobalConfigSpec {
	v2 := NewDefaultConfig()

	v2.Server = v1.Server
	v2.Kafka = v1.Kafka
	v2.KvStore = v1.KvStore
	v2.Tls.UseTls = v1.Tls.UseTls
	v2.Tls.CACert = v1.Tls.CACert
	v2.Tls.Cert = v1.Tls.Cert
	v2.Tls.Key = v1.Tls.Key
	if v1.Tls.Verify != "" {
		if b, err := strconv.ParseBool(v1.Tls.Verify); err == nil {
			v2.Tls.Verify = b
		}
	}
	v2.Grpc.Timeout = v1.Grpc.Timeout
	v2.Grpc.MaxCallRecvMsgSize = v1.Grpc.MaxCallRecvMsgSize
	v2.KvStoreConfig.Timeout = v1.KvStoreConfig.Timeout
	v2.K8sConfig = v1.K8sConfig
	return v2
}
