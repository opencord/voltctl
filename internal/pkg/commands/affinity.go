/*
 * Copyright 2019-present Ciena Corporation
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
package commands

import (
	"context"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/opencord/voltctl/pkg/format"
)

type GoRoutineCountOutput struct {
	Count uint32
}

type GoRoutineCountOpts struct {
	OutputOptions
}

type ResultOutput struct {
	Status string
	Error  string
}

type SetLogLevelOpts struct {
	OutputOptions
	Component string `short:"c" long:"component" description:"Component name to set filter level"`
	Package   string `short:"p" long:"package" description:"Package name to set filter level"`
	Args      struct {
		Level string
	} `positional-args:"yes" required:"yes"`
}

type AffinityOpts struct {
	GoRoutineCount GoRoutineCountOpts `command:"goroutinecount"`
	SetLogLevel    SetLogLevelOpts    `command:"setloglevel"`
}

var affinityOpts = AffinityOpts{}

var goroutinecountInfo = GoRoutineCountOutput{}

var resultInfo = ResultOutput{}

const DefaultGoRoutineCountFormat = `{{.Count}}`

const DefaultResultFormat = "{{.Status}}{{ .Error }}"

func RegisterAffinityCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("affinity", "affinity router stuff", "Affinity Router Stuff", &affinityOpts)
	if err != nil {
		panic(err)
	}
}

func (options *GoRoutineCountOpts) Execute(args []string) error {
	conn, err := NewAffinityConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	descriptor, method, err := GetMethod("get-goroutine-count")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
	if err != nil {
		return err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}
	count, err := d.TryGetFieldByName("count")
	if err != nil {
		return err
	}

	goroutinecountInfo.Count = count.(uint32)

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DefaultGoRoutineCountFormat
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      goroutinecountInfo,
	}

	GenerateOutput(&result)
	return nil
}

func (options *SetLogLevelOpts) Execute(args []string) error {
	conn, err := NewAffinityConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	descriptor, method, err := GetMethod("set-log-level")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	ll := make(map[string]interface{})
	ll["component_name"] = options.Component
	ll["package_name"] = options.Package
	ll["level"] = options.Args.Level

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{"afrouter.LogLevel": ll},
	}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
	if err != nil {
		return err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}
	success, err := d.TryGetFieldByName("success")
	if err != nil {
		return err
	}

	if success.(bool) {
		resultInfo.Status = "Success"
	} else {
		resultInfo.Status = "Failure"
	}

	errorText, err := d.TryGetFieldByName("error")
	if err != nil {
		return err
	}
	resultInfo.Error = errorText.(string)

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DefaultResultFormat
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      resultInfo,
	}

	GenerateOutput(&result)
	return nil
}
