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
