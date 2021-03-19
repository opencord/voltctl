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

type GrpcConfigSpec struct {
	Timeout            time.Duration `yaml:"timeout"`
	MaxCallRecvMsgSize string        `yaml:"maxCallRecvMsgSize"`
}

type KvStoreConfigSpec struct {
	Timeout time.Duration `yaml:"timeout"`
}

type TlsConfigSpec struct {
	UseTls bool   `yaml:"useTls"`
	CACert string `yaml:"caCert"`
	Cert   string `yaml:"cert"`
	Key    string `yaml:"key"`
	Verify string `yaml:"verify"`
}

type GlobalConfigSpec struct {
	Server        string            `yaml:"server"`
	Kafka         string            `yaml:"kafka"`
	KvStore       string            `yaml:"kvstore"`
	Tls           TlsConfigSpec     `yaml:"tls"`
	Grpc          GrpcConfigSpec    `yaml:"grpc"`
	KvStoreConfig KvStoreConfigSpec `yaml:"kvstoreconfig"`
	K8sConfig     string            `yaml:"-"`
}
