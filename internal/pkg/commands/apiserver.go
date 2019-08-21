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

/* TODO

   This was a proof-of-concept making voltctl talk to api-server.

   See if GetGoRoutineCount is useful ... If not, delete.
*/

type GoRoutineCountOutput struct {
	Count uint32
}

type GoRoutineCountOpts struct {
	OutputOptions
}

type ApiServerOpts struct {
	GoRoutineCount GoRoutineCountOpts `command:"goroutinecount"`
}

var apiServerOpts = ApiServerOpts{}

var goroutinecountInfo = GoRoutineCountOutput{}

const (
	DEFAULT_GOROUTINECOUNT_FORMAT = `{{.Count}}`
)

func RegisterApiServerCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("apiserver", "api-server stuff", "Api-Server Stuff", &apiServerOpts)
	if err != nil {
		panic(err)
	}
}

func (options *GoRoutineCountOpts) Execute(args []string) error {
	conn, err := NewConnection()
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
		outputFormat = DEFAULT_GOROUTINECOUNT_FORMAT
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
