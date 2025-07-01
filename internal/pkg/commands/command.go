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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb" //nolint:staticcheck
	"github.com/golang/protobuf/proto"  //nolint:staticcheck
	configv1 "github.com/opencord/voltctl/internal/pkg/apis/config/v1"
	configv2 "github.com/opencord/voltctl/internal/pkg/apis/config/v2"
	configv3 "github.com/opencord/voltctl/internal/pkg/apis/config/v3"
	"github.com/opencord/voltctl/pkg/filter"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/order"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	yaml "gopkg.in/yaml.v2"
)

type OutputType uint8

const (
	OUTPUT_TABLE OutputType = iota
	OUTPUT_JSON
	OUTPUT_YAML

	supportedKvStoreType = "etcd"
)

var (
	ParamNames = map[string]map[string]string{
		"v1": {
			"ID": "voltha.ID",
		},
		"v2": {
			"ID": "common.ID",
		},
		"v3": {
			"ID":             "common.ID",
			"port":           "voltha.Port",
			"ValueSpecifier": "common.ValueSpecifier",
		},
	}

	CharReplacer = strings.NewReplacer("\\t", "\t", "\\n", "\n")

	GlobalConfig = configv3.NewDefaultConfig()

	GlobalCommandOptions = make(map[string]map[string]string)

	GlobalOptions struct {
		Config  string `short:"c" long:"config" env:"VOLTCONFIG" value-name:"FILE" default:"" description:"Location of client config file"`
		Stack   string `short:"v" long:"stack" env:"STACK" value-name:"STACK" default:"" description:"Name of stack to use in multistack deployment"`
		Server  string `short:"s" long:"server" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of VOLTHA"`
		Kafka   string `short:"k" long:"kafka" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of Kafka"`
		KvStore string `short:"e" long:"kvstore" env:"KVSTORE" value-name:"SERVER:PORT" description:"IP/Host and port of KV store (etcd)"`

		// nolint: staticcheck
		Debug              bool   `short:"d" long:"debug" description:"Enable debug mode"`
		Timeout            string `short:"t" long:"timeout" description:"API call timeout duration" value-name:"DURATION" default:""`
		UseTLS             bool   `long:"tls" description:"Use TLS"`
		CACert             string `long:"tlscacert" value-name:"CA_CERT_FILE" description:"Trust certs signed only by this CA"`
		Cert               string `long:"tlscert" value-name:"CERT_FILE" description:"Path to TLS vertificate file"`
		Key                string `long:"tlskey" value-name:"KEY_FILE" description:"Path to TLS key file"`
		Verify             bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
		K8sConfig          string `short:"8" long:"k8sconfig" env:"KUBECONFIG" value-name:"FILE" default:"" description:"Location of Kubernetes config file"`
		KvStoreTimeout     string `long:"kvstoretimeout" env:"KVSTORE_TIMEOUT" value-name:"DURATION" default:"" description:"timeout for calls to KV store"`
		CommandOptions     string `short:"o" long:"command-options" env:"VOLTCTL_COMMAND_OPTIONS" value-name:"FILE" default:"" description:"Location of command options default configuration file"`
		MaxCallRecvMsgSize string `short:"m" long:"maxcallrecvmsgsize" description:"Max GRPC Client request size limit in bytes (eg: 4MB)" value-name:"SIZE" default:"4M"`
	}

	Debug = log.New(os.Stdout, "DEBUG: ", 0)
	Info  = log.New(os.Stdout, "INFO: ", 0)
	Warn  = log.New(os.Stderr, "WARN: ", 0)
	Error = log.New(os.Stderr, "ERROR: ", 0)
)

type OutputOptions struct {
	Format string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	Quiet  bool   `short:"q" long:"quiet" description:"Output only the IDs of the objects"`
	// nolint: staticcheck
	OutputAs  string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
	NameLimit int    `short:"l" long:"namelimit" default:"-1" value-name:"LIMIT" description:"Limit the depth (length) in the table column name"`
}

type FlowIdOptions struct {
	HexId bool `short:"x" long:"hex-id" description:"Output Ids in hex format"`
}

type GroupListOptions struct {
	Bucket bool `short:"b" long:"buckets" description:"Display Buckets"`
}
type ListOutputOptions struct {
	OutputOptions
	Filter  string `short:"f" long:"filter" default:"" value-name:"FILTER" description:"Only display results that match filter"`
	OrderBy string `short:"r" long:"orderby" default:"" value-name:"ORDER" description:"Specify the sort order of the results"`
}

type OutputOptionsJson struct {
	Format string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	Quiet  bool   `short:"q" long:"quiet" description:"Output only the IDs of the objects"`
	// nolint: staticcheck
	OutputAs  string `short:"o" long:"outputas" default:"json" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
	NameLimit int    `short:"l" long:"namelimit" default:"-1" value-name:"LIMIT" description:"Limit the depth (length) in the table column name"`
}

type ListOutputOptionsJson struct {
	OutputOptionsJson
	Filter  string `short:"f" long:"filter" default:"" value-name:"FILTER" description:"Only display results that match filter"`
	OrderBy string `short:"r" long:"orderby" default:"" value-name:"ORDER" description:"Specify the sort order of the results"`
}

func toOutputType(in string) OutputType {
	switch in {
	case "table":
		fallthrough
	default:
		return OUTPUT_TABLE
	case "json":
		return OUTPUT_JSON
	case "yaml":
		return OUTPUT_YAML
	}
}

type CommandResult struct {
	Format    format.Format
	Filter    string
	OrderBy   string
	OutputAs  OutputType
	NameLimit int
	Data      interface{}
}

func GetCommandOptionWithDefault(name, option, defaultValue string) string {
	if cmd, ok := GlobalCommandOptions[name]; ok {
		if val, ok := cmd[option]; ok {
			return CharReplacer.Replace(val)
		}
	}
	return defaultValue
}

var sizeParser = regexp.MustCompile(`^([0-9]+)([PETGMK]?)I?B?$`)

func parseSize(size string) (uint64, error) {

	parts := sizeParser.FindAllStringSubmatch(strings.ToUpper(size), -1)
	if len(parts) == 0 {
		return 0, fmt.Errorf("size: invalid size '%s'", size)
	}
	value, err := strconv.ParseUint(parts[0][1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("size: invalid size '%s'", size)
	}
	switch parts[0][2] {
	case "E":
		value = value * 1024 * 1024 * 1024 * 1024 * 1024 * 1024
	case "P":
		value = value * 1024 * 1024 * 1024 * 1024 * 1024
	case "T":
		value = value * 1024 * 1024 * 1024 * 1024
	case "G":
		value = value * 1024 * 1024 * 1024
	case "M":
		value = value * 1024 * 1024
	case "K":
		value = value * 1024
	default:
	}
	return value, nil
}

func ProcessGlobalOptions() {
	ReadConfig()

	// If a stack is selected via command line set it
	if GlobalOptions.Stack != "" {
		if GlobalConfig.StackByName(GlobalOptions.Stack) == nil {
			Error.Fatalf("stack specified, '%s', not found in configuration",
				GlobalOptions.Stack)
		}
		GlobalConfig.CurrentStack = GlobalOptions.Stack
	}

	ApplyOptionOverrides(GlobalConfig.Current())

	// If a k8s cert/key were not specified, then attempt to read it from
	// any $HOME/.kube/config if it exists
	if len(GlobalOptions.K8sConfig) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			Warn.Printf("Unable to discover the user's home directory: %s", err)
			home = "~"
		}
		GlobalOptions.K8sConfig = filepath.Join(home, ".kube", "config")
	}

	if len(GlobalOptions.CommandOptions) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			Warn.Printf("Unable to discover the user's home directory: %s", err)
			home = "~"
		}
		GlobalOptions.CommandOptions = filepath.Join(home, ".volt", "command_options")
	}

	if info, err := os.Stat(GlobalOptions.CommandOptions); err == nil && !info.IsDir() {
		optionsFile, err := os.ReadFile(GlobalOptions.CommandOptions)
		if err != nil {
			Error.Fatalf("Unable to read command options configuration file '%s' : %s",
				GlobalOptions.CommandOptions, err.Error())
		}
		if err = yaml.Unmarshal(optionsFile, &GlobalCommandOptions); err != nil {
			Error.Fatalf("Unable to parse the command line options configuration file '%s': %s",
				GlobalOptions.CommandOptions, err.Error())
		}
	}
}

func ApplyOptionOverrides(stack *configv3.StackConfigSpec) {

	if stack == nil {
		// nothing to do
		return
	}
	// Override from command line
	if GlobalOptions.Server != "" {
		stack.Server = GlobalOptions.Server
	}

	if GlobalOptions.UseTLS {
		stack.Tls.UseTls = true
	}

	if GlobalOptions.Verify {
		stack.Tls.Verify = true
	}

	if GlobalOptions.Kafka != "" {
		stack.Kafka = GlobalOptions.Kafka
	}

	if GlobalOptions.KvStore != "" {
		stack.KvStore = GlobalOptions.KvStore
	}

	if GlobalOptions.KvStoreTimeout != "" {
		timeout, err := time.ParseDuration(GlobalOptions.KvStoreTimeout)
		if err != nil {
			Error.Fatalf("Unable to parse specified KV strore timeout duration '%s': %s",
				GlobalOptions.KvStoreTimeout, err.Error())
		}
		stack.KvStoreConfig.Timeout = timeout
	}

	if GlobalOptions.Timeout != "" {
		timeout, err := time.ParseDuration(GlobalOptions.Timeout)
		if err != nil {
			Error.Fatalf("Unable to parse specified timeout duration '%s': %s",
				GlobalOptions.Timeout, err.Error())
		}
		stack.Grpc.Timeout = timeout
	}

	if GlobalOptions.MaxCallRecvMsgSize != "" {
		stack.Grpc.MaxCallRecvMsgSize = GlobalOptions.MaxCallRecvMsgSize
	}
}

func homeDir() (string, error) {

	// First attempt is using the current user information, if that
	// fails we attempt to use the environment variables via a call
	// to os.UserHomeDir()
	cur, err := user.Current()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("unable to get current user or determine home directory via environment")
		}
		return home, nil
	}
	if len(cur.HomeDir) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.New("unable to get determine home directory via user information or environment")
		}
		return home, nil
	}
	return cur.HomeDir, nil
}

func ReadConfig() {
	// assume that the config file was specified until it is
	// determined it was not
	if len(GlobalOptions.Config) == 0 {
		home, err := homeDir()
		if err != nil {
			Error.Fatalf("Unable to create default configuration file path: %s", err.Error())
		}
		GlobalOptions.Config = filepath.Join(home, ".volt", "config")
	}

	// If the config file value is `~` or begins with `~/` then
	// attempt to expand that to the user's home directory
	if GlobalOptions.Config == "~" {
		home, err := homeDir()
		if err != nil {
			Error.Fatalf("Unable to expand config file path '%s': %s",
				GlobalOptions.Config, err)
		}
		GlobalOptions.Config = home
	} else if strings.HasPrefix(GlobalOptions.Config, "~/") {
		home, err := homeDir()
		if err != nil {
			Error.Fatalf("Unable to expand config file path '%s': %s",
				GlobalOptions.Config, err)
		}
		GlobalOptions.Config = filepath.Join(home,
			GlobalOptions.Config[2:])
	}

	if info, err := os.Stat(GlobalOptions.Config); err == nil && !info.IsDir() {
		configFile, err := os.ReadFile(GlobalOptions.Config)
		if err != nil {
			Error.Fatalf("Unable to read the configuration file '%s': %s",
				GlobalOptions.Config, err.Error())
		}
		// First try the latest version of the config api then work
		// backwards
		if err = yaml.Unmarshal(configFile, &GlobalConfig); err != nil {
			GlobalConfigV2 := configv2.NewDefaultConfig()
			if err = yaml.Unmarshal(configFile, &GlobalConfigV2); err != nil {
				GlobalConfigV1 := configv1.NewDefaultConfig()
				if err = yaml.Unmarshal(configFile, &GlobalConfigV1); err != nil {
					Error.Fatalf("Unable to parse the configuration file '%s': %s",
						GlobalOptions.Config, err.Error())
				}
				GlobalConfig = configv3.FromConfigV1(GlobalConfigV1)
			} else {
				GlobalConfig = configv3.FromConfigV2(GlobalConfigV2)
			}
		}
	}
}

func NewConnection() (*grpc.ClientConn, error) {
	ProcessGlobalOptions()

	// convert grpc.msgSize into bytes
	n, err := parseSize(GlobalConfig.Current().Grpc.MaxCallRecvMsgSize)
	if err != nil {
		Error.Fatalf("Cannot convert msgSize %s to bytes", GlobalConfig.Current().Grpc.MaxCallRecvMsgSize)
	}

	var opts []grpc.DialOption

	opts = append(opts,
		grpc.WithDisableRetry(),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(n))))

	if GlobalConfig.Current().Tls.UseTls {
		creds := credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: !GlobalConfig.Current().Tls.Verify})
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}
	ctx, cancel := context.WithTimeout(context.TODO(),
		GlobalConfig.Current().Grpc.ConnectTimeout)
	defer cancel()
	return grpc.DialContext(ctx, GlobalConfig.Current().Server, opts...)
}

func ConvertJsonProtobufArray(data_in interface{}) (string, error) {
	result := ""

	slice := reflect.ValueOf(data_in)
	if slice.Kind() != reflect.Slice {
		return "", errors.New("Not a slice")
	}

	result = result + "["

	marshaler := jsonpb.Marshaler{EmitDefaults: true}
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		protoMessage, okay := item.(proto.Message)
		if !okay {
			return "", errors.New("Failed to convert item to a proto.Message")
		}
		asJson, err := marshaler.MarshalToString(protoMessage)
		if err != nil {
			return "", fmt.Errorf("Failed to marshal the json: %s", err)
		}

		result = result + asJson

		if i < slice.Len()-1 {
			result = result + ","
		}
	}

	result = result + "]"

	return result, nil
}

func GenerateOutput(result *CommandResult) {
	if result != nil && result.Data != nil {
		data := result.Data
		if result.Filter != "" {
			f, err := filter.Parse(result.Filter)
			if err != nil {
				Error.Fatalf("Unable to parse specified output filter '%s': %s", result.Filter, err.Error())
			}
			data, err = f.Process(data)
			if err != nil {
				Error.Fatalf("Unexpected error while filtering command results: %s", err.Error())
			}
		}
		if result.OrderBy != "" {
			s, err := order.Parse(result.OrderBy)
			if err != nil {
				Error.Fatalf("Unable to parse specified sort specification '%s': %s", result.OrderBy, err.Error())
			}
			data, err = s.Process(data)
			if err != nil {
				Error.Fatalf("Unexpected error while sorting command result: %s", err.Error())
			}
		}
		if result.OutputAs == OUTPUT_TABLE {
			tableFormat := format.Format(result.Format)
			if err := tableFormat.Execute(os.Stdout, true, result.NameLimit, data); err != nil {
				Error.Fatalf("Unexpected error while attempting to format results as table : %s", err.Error())
			}
		} else if result.OutputAs == OUTPUT_JSON {
			// first try to convert it as an array of protobufs
			asJson, err := ConvertJsonProtobufArray(data)
			if err != nil {
				// if that fails, then just do a standard json conversion
				asJsonB, err := json.Marshal(&data)
				if err != nil {
					Error.Fatalf("Unexpected error while processing command results to JSON: %s", err.Error())
				}
				asJson = string(asJsonB)
			}
			fmt.Printf("%s", asJson)
		} else if result.OutputAs == OUTPUT_YAML {
			asYaml, err := yaml.Marshal(&data)
			if err != nil {
				Error.Fatalf("Unexpected error while processing command results to YAML: %s", err.Error())
			}
			fmt.Printf("%s", asYaml)
		}
	}
}
