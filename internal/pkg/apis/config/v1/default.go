package config

import (
	"github.com/opencord/voltctl/internal/pkg/apis/config"
)

func NewDefaultConfig() *GlobalConfigSpec {
	return &GlobalConfigSpec{
		Server:  "localhost:55555",
		Kafka:   "localhost:9093",
		KvStore: "localhost:2379",
		Tls: TlsConfigSpec{
			UseTls: false,
			Verify: "",
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
