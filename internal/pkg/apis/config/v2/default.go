package config

import (
	"github.com/opencord/voltctl/internal/pkg/apis/config"
)

func NewDefaultConfig() *GlobalConfigSpec {
	return &GlobalConfigSpec{
		ApiVersion:   "v2",
		CurrentStack: "default",
		Stacks:       []*StackConfigSpec{NewDefaultStack("default")},
	}
}

func NewDefaultStack(name string) *StackConfigSpec {
	return &StackConfigSpec{
		Name:    name,
		Server:  "localhost:55555",
		Kafka:   "localhost:9093",
		KvStore: "localhost:2379",
		Tls: TlsConfigSpec{
			UseTls: false,
			Verify: false,
		},
		Grpc: GrpcConfigSpec{
			Timeout:            config.DefaultGrpcTimeout,
			MaxCallRecvMsgSize: config.DefaultGrpcMaxCallRecvMsgSize,
		},
		KvStoreConfig: KvStoreConfigSpec{
			Timeout: config.DefaultKvTimeout,
		},
	}
}
