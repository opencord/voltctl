package config

import "time"

const (
	DefaultApiHost = "localhost"
	DefaultApiPort = 55555

	DefaultKafkaHost = "localhost"
	DefaultKafkaPort = 9092

	SupportedKvStoreType = "etcd"
	DefaultKvHost        = "localhost"
	DefaultKvPort        = 2379
	DefaultKvTimeout     = time.Second * 5

	DefaultGrpcTimeout            = time.Minute * 5
	DefaultGrpcMaxCallRecvMsgSize = "4MB"
)
