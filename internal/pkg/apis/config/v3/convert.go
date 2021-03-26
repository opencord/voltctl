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
	configv2 "github.com/opencord/voltctl/internal/pkg/apis/config/v2"
)

func FromConfigV1(v1 *configv1.GlobalConfigSpec) *GlobalConfigSpec {
	v3 := NewDefaultConfig()
	s3 := v3.Current()

	s3.Server = v1.Server
	s3.Kafka = v1.Kafka
	s3.KvStore = v1.KvStore
	s3.Tls.UseTls = v1.Tls.UseTls
	s3.Tls.CACert = v1.Tls.CACert
	s3.Tls.Cert = v1.Tls.Cert
	s3.Tls.Key = v1.Tls.Key
	if v1.Tls.Verify != "" {
		if b, err := strconv.ParseBool(v1.Tls.Verify); err == nil {
			s3.Tls.Verify = b
		}
	}
	s3.Grpc.Timeout = v1.Grpc.Timeout
	s3.Grpc.MaxCallRecvMsgSize = v1.Grpc.MaxCallRecvMsgSize
	s3.KvStoreConfig.Timeout = v1.KvStoreConfig.Timeout
	s3.K8sConfig = v1.K8sConfig
	return v3
}

func FromConfigV2(v2 *configv2.GlobalConfigSpec) *GlobalConfigSpec {
	v3 := NewDefaultConfig()
	s3 := v3.Current()

	s3.Server = v2.Server
	s3.Kafka = v2.Kafka
	s3.KvStore = v2.KvStore
	s3.Tls.UseTls = v2.Tls.UseTls
	s3.Tls.CACert = v2.Tls.CACert
	s3.Tls.Cert = v2.Tls.Cert
	s3.Tls.Key = v2.Tls.Key
	s3.Tls.Verify = v2.Tls.Verify
	s3.Grpc.ConnectTimeout = v2.Grpc.ConnectTimeout
	s3.Grpc.Timeout = v2.Grpc.Timeout
	s3.Grpc.MaxCallRecvMsgSize = v2.Grpc.MaxCallRecvMsgSize
	s3.KvStoreConfig.Timeout = v2.KvStoreConfig.Timeout
	s3.K8sConfig = v2.K8sConfig
	return v3
}
