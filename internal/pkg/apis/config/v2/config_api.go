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
	"errors"
	"log"
	"os"
	"time"

	"github.com/opencord/voltctl/internal/pkg/apis/config"
	yaml "gopkg.in/yaml.v2"
)

func (c *GlobalConfigSpec) AddStack(name string) {
	for _, stack := range c.Stacks {
		if stack.Name == name {
			return
		}
	}

	c.Stacks = append(c.Stacks, NewDefaultStack(name))
}

func (c *GlobalConfigSpec) GetStacks() []config.StackConfigAPI {
	var stacks []config.StackConfigAPI
	for _, stack := range c.Stacks {
		stacks = append(stacks, stack)
	}
	return stacks
}

func (c *GlobalConfigSpec) GetStackByName(name string) config.StackConfigAPI {
	for _, stack := range c.Stacks {
		if stack.Name == name {
			return stack
		}
	}
	return nil
}

func (c *GlobalConfigSpec) GetCurrentAsStack() config.StackConfigAPI {
	stack := c.GetStackByName(c.GetCurrentStack())
	// If no stack found, error out
	if stack == nil {
		log.New(os.Stderr, "ERROR: ", 0).
			Fatalf("Unknown or no stack name specified, have '%s'", c.CurrentStack)
	}
	return stack
}

func (c *GlobalConfigSpec) GetCurrentStack() string {
	return c.CurrentStack
}

func (c *GlobalConfigSpec) SetCurrentStack(name string) error {
	for _, stack := range c.Stacks {
		if stack.Name == name {
			c.CurrentStack = name
			return nil
		}
	}
	return errors.New("not-found")
}

func (c *GlobalConfigSpec) DeleteStack(name string) error {
	for i, stack := range c.Stacks {
		if stack.Name == name {
			if c.CurrentStack == name {
				c.CurrentStack = ""
			}
			c.Stacks = append(c.Stacks[:i], c.Stacks[i+1:]...)
			return nil
		}
	}
	return errors.New("not-found")
}

func (c *GlobalConfigSpec) Write(name string) error {
	w, err := os.OpenFile(name, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer w.Close()
	encode := yaml.NewEncoder(w)
	if err := encode.Encode(c); err != nil {
		return err
	}
	return nil
}

func (c *GlobalConfigSpec) GetName() string {
	return c.GetCurrentAsStack().GetName()
}

func (c *GlobalConfigSpec) SetName(name string) {
	c.GetCurrentAsStack().SetName(name)
}

func (c *GlobalConfigSpec) GetServer() string {
	return c.GetCurrentAsStack().GetServer()
}

func (c *GlobalConfigSpec) SetServer(server string) {
	c.GetCurrentAsStack().SetServer(server)
}

func (c *GlobalConfigSpec) GetKafka() string {
	return c.GetCurrentAsStack().GetKafka()
}

func (c *GlobalConfigSpec) SetKafka(kafka string) {
	c.GetCurrentAsStack().SetKafka(kafka)
}

func (c *GlobalConfigSpec) GetKvStore() string {
	return c.GetCurrentAsStack().GetKvStore()
}

func (c *GlobalConfigSpec) SetKvStore(store string) {
	c.GetCurrentAsStack().SetKvStore(store)
}

func (c *GlobalConfigSpec) GetKvStoreTimeout() time.Duration {
	return c.GetCurrentAsStack().GetKvStoreTimeout()
}

func (c *GlobalConfigSpec) SetKvStoreTimeout(t time.Duration) {
	c.GetCurrentAsStack().SetKvStoreTimeout(t)
}

func (c *GlobalConfigSpec) GetGrpcTimeout() time.Duration {
	return c.GetCurrentAsStack().GetGrpcTimeout()
}

func (c *GlobalConfigSpec) SetGrpcTimeout(t time.Duration) {
	c.GetCurrentAsStack().SetGrpcTimeout(t)
}

func (c *GlobalConfigSpec) GetGrpcMaxCallRecvMsgSize() string {
	return c.GetCurrentAsStack().GetGrpcMaxCallRecvMsgSize()
}

func (c *GlobalConfigSpec) SetGrpcMaxCallRecvMsgSize(size string) {
	c.GetCurrentAsStack().SetGrpcMaxCallRecvMsgSize(size)
}
