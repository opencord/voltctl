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
	"fmt"
	"github.com/opencord/voltctl/pkg/filter"
	"github.com/opencord/voltctl/pkg/format"
	"github.com/opencord/voltctl/pkg/order"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type OutputType uint8

const (
	OUTPUT_TABLE OutputType = iota
	OUTPUT_JSON
	OUTPUT_YAML
)

type GrpcConfigSpec struct {
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
	ApiVersion string         `yaml:"apiVersion"`
	Server     string         `yaml:"server"`
	Kafka      string         `yaml:"kafka"`
	Tls        TlsConfigSpec  `yaml:"tls"`
	Grpc       GrpcConfigSpec `yaml:"grpc"`
	K8sConfig  string         `yaml:"-"`
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
			"ID":   "common.ID",
			"port": "voltha.Port",
		},
	}

	CharReplacer = strings.NewReplacer("\\t", "\t", "\\n", "\n")

	GlobalConfig = GlobalConfigSpec{
		ApiVersion: "v3",
		Server:     "localhost:55555",
		Kafka:      "",
		Tls: TlsConfigSpec{
			UseTls: false,
		},
		Grpc: GrpcConfigSpec{
			Timeout: time.Minute * 5,
		},
	}

	GlobalCommandOptions = make(map[string]map[string]string)

	GlobalOptions struct {
		Config string `short:"c" long:"config" env:"VOLTCONFIG" value-name:"FILE" default:"" description:"Location of client config file"`
		Server string `short:"s" long:"server" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of VOLTHA"`
		Kafka  string `short:"k" long:"kafka" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of Kafka"`
		// Do not set the default for the API version here, else it will override the value read in the config
		// nolint: staticcheck
		ApiVersion     string `short:"a" long:"apiversion" description:"API version" value-name:"VERSION" choice:"v1" choice:"v2" choice:"v3"`
		Debug          bool   `short:"d" long:"debug" description:"Enable debug mode"`
		Timeout        string `short:"t" long:"timeout" description:"API call timeout duraction" value-name:"DURATION" default:""`
		UseTLS         bool   `long:"tls" description:"Use TLS"`
		CACert         string `long:"tlscacert" value-name:"CA_CERT_FILE" description:"Trust certs signed only by this CA"`
		Cert           string `long:"tlscert" value-name:"CERT_FILE" description:"Path to TLS vertificate file"`
		Key            string `long:"tlskey" value-name:"KEY_FILE" description:"Path to TLS key file"`
		Verify         bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
		K8sConfig      string `short:"8" long:"k8sconfig" env:"KUBECONFIG" value-name:"FILE" default:"" description:"Location of Kubernetes config file"`
		CommandOptions string `short:"o" long:"command-options" env:"VOLTCTL_COMMAND_OPTIONS" value-name:"FILE" default:"" description:"Location of command options default configuration file"`
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
	if GlobalOptions.Kafka != "" {
		GlobalConfig.Kafka = GlobalOptions.Kafka
	}
	if GlobalOptions.ApiVersion != "" {
		GlobalConfig.ApiVersion = GlobalOptions.ApiVersion
	}

	if GlobalOptions.Timeout != "" {
		timeout, err := time.ParseDuration(GlobalOptions.Timeout)
		if err != nil {
			Error.Fatalf("Unable to parse specified timeout duration '%s': %s",
				GlobalOptions.Timeout, err.Error())
		}
		GlobalConfig.Grpc.Timeout = timeout
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
	return grpc.Dial(GlobalConfig.Server, grpc.WithInsecure())
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
			asJson, err := json.Marshal(&data)
			if err != nil {
				Error.Fatalf("Unexpected error while processing command results to JSON: %s", err.Error())
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
