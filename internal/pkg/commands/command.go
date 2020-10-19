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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/opencord/voltctl/pkg/filter"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/order"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type OutputType uint8

const (
	OUTPUT_TABLE OutputType = iota
	OUTPUT_JSON
	OUTPUT_YAML

	defaultApiHost = "localhost"
	defaultApiPort = 55555

	defaultKafkaHost = "localhost"
	defaultKafkaPort = 9092

	supportedKvStoreType = "etcd"
	defaultKvHost        = "localhost"
	defaultKvPort        = 2379
	defaultKvTimeout     = time.Second * 5

	defaultGrpcTimeout            = time.Minute * 5
	defaultGrpcMaxCallRecvMsgSize = "4MB"
)

type GrpcConfigSpec struct {
	Timeout            time.Duration `yaml:"timeout"`
	MaxCallRecvMsgSize string        `yaml:"maxCallRecvMsgSize"`
}

type KvStoreConfigSpec struct {
	Timeout time.Duration `yaml:"timeout"`
}

type TlsConfigSpec struct {
	UseTls bool   `yaml:"useTls"`
	CACert string `yaml:"caCert"`
	Cert   string `yaml:"cert"`
	Key    string `yaml:"key"`
	Verify string `yaml:"verify"`
}

type GlobalConfigSpec struct {
	Server        string            `yaml:"server"`
	Kafka         string            `yaml:"kafka"`
	KvStore       string            `yaml:"kvstore"`
	Tls           TlsConfigSpec     `yaml:"tls"`
	Grpc          GrpcConfigSpec    `yaml:"grpc"`
	KvStoreConfig KvStoreConfigSpec `yaml:"kvstoreconfig"`
	K8sConfig     string            `yaml:"-"`
}

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

	GlobalConfig = GlobalConfigSpec{
		Server:  "localhost:55555",
		Kafka:   "",
		KvStore: "localhost:2379",
		Tls: TlsConfigSpec{
			UseTls: false,
		},
		Grpc: GrpcConfigSpec{
			Timeout:            defaultGrpcTimeout,
			MaxCallRecvMsgSize: defaultGrpcMaxCallRecvMsgSize,
		},
		KvStoreConfig: KvStoreConfigSpec{
			Timeout: defaultKvTimeout,
		},
	}

	GlobalCommandOptions = make(map[string]map[string]string)

	GlobalOptions struct {
		Config  string `short:"c" long:"config" env:"VOLTCONFIG" value-name:"FILE" default:"" description:"Location of client config file"`
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

type ListOutputOptions struct {
	OutputOptions
	Filter  string `short:"f" long:"filter" default:"" value-name:"FILTER" description:"Only display results that match filter"`
	OrderBy string `short:"r" long:"orderby" default:"" value-name:"ORDER" description:"Specify the sort order of the results"`
}

type UpdateAdditionalOptions struct {
	Wide bool `short:"w" long:"wide" description:"Display additional fields"`
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

func splitEndpoint(ep, defaultHost string, defaultPort int) (string, int, error) {
	port := defaultPort
	host, sPort, err := net.SplitHostPort(ep)
	if err != nil {
		if addrErr, ok := err.(*net.AddrError); ok {
			if addrErr.Err != "missing port in address" {
				return "", 0, err
			}
			host = ep
		} else {
			return "", 0, err
		}
	} else if len(strings.TrimSpace(sPort)) > 0 {
		val, err := strconv.Atoi(sPort)
		if err != nil {
			return "", 0, err
		}
		port = val
	}
	if len(strings.TrimSpace(host)) == 0 {
		host = defaultHost
	}
	return strings.Trim(host, "]["), port, nil
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
	if len(GlobalOptions.Config) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			Warn.Printf("Unable to discover the user's home directory: %s", err)
			home = "~"
		}
		GlobalOptions.Config = filepath.Join(home, ".volt", "config")
	}

	if info, err := os.Stat(GlobalOptions.Config); err == nil && !info.IsDir() {
		configFile, err := ioutil.ReadFile(GlobalOptions.Config)
		if err != nil {
			Error.Fatalf("Unable to read the configuration file '%s': %s",
				GlobalOptions.Config, err.Error())
		}
		if err = yaml.Unmarshal(configFile, &GlobalConfig); err != nil {
			Error.Fatalf("Unable to parse the configuration file '%s': %s",
				GlobalOptions.Config, err.Error())
		}
	}

	// Override from command line
	if GlobalOptions.Server != "" {
		GlobalConfig.Server = GlobalOptions.Server
	}
	host, port, err := splitEndpoint(GlobalConfig.Server, defaultApiHost, defaultApiPort)
	if err != nil {
		Error.Fatalf("voltha API endport incorrectly specified '%s':%s",
			GlobalConfig.Server, err)
	}
	GlobalConfig.Server = net.JoinHostPort(host, strconv.Itoa(port))

	if GlobalOptions.Kafka != "" {
		GlobalConfig.Kafka = GlobalOptions.Kafka
	}
	host, port, err = splitEndpoint(GlobalConfig.Kafka, defaultKafkaHost, defaultKafkaPort)
	if err != nil {
		Error.Fatalf("Kafka endport incorrectly specified '%s':%s",
			GlobalConfig.Kafka, err)
	}
	GlobalConfig.Kafka = net.JoinHostPort(host, strconv.Itoa(port))

	if GlobalOptions.KvStore != "" {
		GlobalConfig.KvStore = GlobalOptions.KvStore
	}
	host, port, err = splitEndpoint(GlobalConfig.KvStore, defaultKvHost, defaultKvPort)
	if err != nil {
		Error.Fatalf("KV store endport incorrectly specified '%s':%s",
			GlobalConfig.KvStore, err)
	}
	GlobalConfig.KvStore = net.JoinHostPort(host, strconv.Itoa(port))

	if GlobalOptions.KvStoreTimeout != "" {
		timeout, err := time.ParseDuration(GlobalOptions.KvStoreTimeout)
		if err != nil {
			Error.Fatalf("Unable to parse specified KV strore timeout duration '%s': %s",
				GlobalOptions.KvStoreTimeout, err.Error())
		}
		GlobalConfig.KvStoreConfig.Timeout = timeout
	}

	if GlobalOptions.Timeout != "" {
		timeout, err := time.ParseDuration(GlobalOptions.Timeout)
		if err != nil {
			Error.Fatalf("Unable to parse specified timeout duration '%s': %s",
				GlobalOptions.Timeout, err.Error())
		}
		GlobalConfig.Grpc.Timeout = timeout
	}

	if GlobalOptions.MaxCallRecvMsgSize != "" {
		GlobalConfig.Grpc.MaxCallRecvMsgSize = GlobalOptions.MaxCallRecvMsgSize
	}

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
		optionsFile, err := ioutil.ReadFile(GlobalOptions.CommandOptions)
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

func NewConnection() (*grpc.ClientConn, error) {
	ProcessGlobalOptions()

	// convert grpc.msgSize into bytes
	n, err := parseSize(GlobalConfig.Grpc.MaxCallRecvMsgSize)
	if err != nil {
		Error.Fatalf("Cannot convert msgSize %s to bytes", GlobalConfig.Grpc.MaxCallRecvMsgSize)
	}

	return grpc.Dial(GlobalConfig.Server, grpc.WithInsecure(), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(n))))
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
