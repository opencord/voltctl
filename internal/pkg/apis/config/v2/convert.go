package config

import (
	"strconv"

	v1config "github.com/opencord/voltctl/internal/pkg/apis/config/v1"
)

func NewFromV1(v1 *v1config.GlobalConfigSpec) *GlobalConfigSpec {
	verify := false

	i := interface{}(v1.Tls.Verify)
	switch v := i.(type) {
	case bool:
		verify = v
	case string:
		// should never have been string
		if v != "" {
			var err error
			if verify, err = strconv.ParseBool(v); err != nil {
				verify = false
			}
		}
	default:
		// ignore
	}
	return &GlobalConfigSpec{
		ApiVersion:   "v2",
		CurrentStack: "default",
		Stacks: []*StackConfigSpec{
			&StackConfigSpec{
				Name:    "default",
				Server:  v1.Server,
				Kafka:   v1.Kafka,
				KvStore: v1.KvStore,
				Tls: TlsConfigSpec{
					UseTls: v1.Tls.UseTls,
					CACert: v1.Tls.CaCert,
					Cert:   v1.Tls.Cert,
					Key:    v1.Tls.Key,
					Verify: verify,
				},
				Grpc: GrpcConfigSpec{
					Timeout:            v1.Grpc.Timeout,
					MaxCallRecvMsgSize: v1.Grpc.MaxCallRecvMsgSize,
				},
				KvStoreConfig: KvStoreConfigSpec{
					Timeout: v1.KvStoreConfig.Timeout,
				},
				K8sConfig: v1.K8sConfig,
			},
		},
	}
}
