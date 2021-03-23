package config

import (
	"time"
)

func (s *StackConfigSpec) GetName() string {
	return s.Name
}

func (s *StackConfigSpec) SetName(name string) {
	s.Name = name
}

func (s *StackConfigSpec) GetServer() string {
	return s.Server
}

func (s *StackConfigSpec) SetServer(server string) {
	s.Server = server
}

func (s *StackConfigSpec) GetKafka() string {
	return s.Kafka
}

func (s *StackConfigSpec) SetKafka(kafka string) {
	s.Kafka = kafka
}

func (s *StackConfigSpec) GetKvStore() string {
	return s.KvStore
}

func (s *StackConfigSpec) SetKvStore(store string) {
	s.KvStore = store
}

func (s *StackConfigSpec) GetKvStoreTimeout() time.Duration {
	return s.KvStoreConfig.Timeout
}

func (s *StackConfigSpec) SetKvStoreTimeout(t time.Duration) {
	s.KvStoreConfig.Timeout = t
}

func (s *StackConfigSpec) GetGrpcTimeout() time.Duration {
	return s.Grpc.Timeout
}

func (s *StackConfigSpec) SetGrpcTimeout(t time.Duration) {
	s.Grpc.Timeout = t
}

func (s *StackConfigSpec) GetGrpcMaxCallRecvMsgSize() string {
	return s.Grpc.MaxCallRecvMsgSize
}

func (s *StackConfigSpec) SetGrpcMaxCallRecvMsgSize(size string) {
	s.Grpc.MaxCallRecvMsgSize = size
}
