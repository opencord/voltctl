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
	"log"
	"os"
	"time"
)

func NewDefaultConfig() *GlobalConfigSpec {
	return &GlobalConfigSpec{
		ApiVersion:   "v3",
		CurrentStack: "",
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
			ConnectTimeout:     5 * time.Second,
			Timeout:            5 * time.Minute,
			MaxCallRecvMsgSize: "4MB",
		},
		KvStoreConfig: KvStoreConfigSpec{
			Timeout: 5 * time.Second,
		},
	}
}

func (g *GlobalConfigSpec) StackByName(name string) *StackConfigSpec {
	for _, stack := range g.Stacks {
		if stack.Name == name {
			return stack
		}
	}
	return nil
}

func (g *GlobalConfigSpec) CurrentAsStack() *StackConfigSpec {
	if g.CurrentStack == "" {
		if len(g.Stacks) == 0 {
			return nil
		}
		return g.Stacks[0]
	}
	return g.StackByName(g.CurrentStack)
}

func (g GlobalConfigSpec) Current() *StackConfigSpec {
	stack := g.CurrentAsStack()
	if stack == nil {
		if len(g.Stacks) > 1 {
			log.New(os.Stderr, "ERROR: ", 0).
				Fatal("multiple stacks configured without current specified")
		}
		log.New(os.Stderr, "ERROR: ", 0).
			Fatalf("current stack specified, '%s', does not exist as a configured stack",
				g.CurrentStack)
	}
	return stack
}
