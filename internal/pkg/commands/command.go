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
	"github.com/ciena/voltctl/pkg/filter"
	"github.com/ciena/voltctl/pkg/format"
	"github.com/ciena/voltctl/pkg/order"
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
	}

	CharReplacer = strings.NewReplacer("\\t", "\t", "\\n", "\n")

	GlobalConfig = GlobalConfigSpec{
		ApiVersion: "v1",
		Server:     "localhost",
		Tls: TlsConfigSpec{
			UseTls: false,
		},
		Grpc: GrpcConfigSpec{
			Timeout: time.Second * 10,
		},
	}

	GlobalOptions struct {
		Config     string `short:"c" long:"config" env:"VOLTCONFIG" value-name:"FILE" default:"" description:"Location of client config file"`
		Server     string `short:"s" long:"server" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of VOLTHA"`
		ApiVersion string `short:"a" long:"apiversion" description:"API version" value-name:"VERSION" choice:"v1" choice:"v2"`
		Debug      bool   `short:"d" long:"debug" description:"Enable debug mode"`
		UseTLS     bool   `long:"tls" description:"Use TLS"`
		CACert     string `long:"tlscacert" value-name:"CA_CERT_FILE" description:"Trust certs signed only by this CA"`
		Cert       string `long:"tlscert" value-name:"CERT_FILE" description:"Path to TLS vertificate file"`
		Key        string `long:"tlskey" value-name:"KEY_FILE" description:"Path to TLS key file"`
		Verify     bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
		K8sConfig  string `short:"8" long:"k8sconfig" env:"KUBECONFIG" value-name:"FILE" default:"" description:"Location of Kubernetes config file"`
	}
)

type OutputOptions struct {
	Format    string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	Quiet     bool   `short:"q" long:"quiet" description:"Output only the IDs of the objects"`
	OutputAs  string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
	NameLimit int    `short:"l" long:"namelimit" default:"-1" value-name:"LIMIT" description:"Limit the depth (length) in the table column name"`
}

type ListOutputOptions struct {
	OutputOptions
	Filter  string `short:"f" long:"filter" default:"" value-name:"FILTER" description:"Only display results that match filter"`
	OrderBy string `short:"r" long:"orderby" default:"" value-name:"ORDER" description:"Specify the sort order of the results"`
}

type OutputOptionsJson struct {
	Format    string `long:"format" value-name:"FORMAT" default:"" description:"Format to use to output structured data"`
	Quiet     bool   `short:"q" long:"quiet" description:"Output only the IDs of the objects"`
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

type config struct {
	ApiVersion string `yaml:"apiVersion"`
	Server     string `yaml:"server"`
}

func ProcessGlobalOptions() {
	if len(GlobalOptions.Config) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Unable to discover they users home directory: %s\n", err)
			home = "~"
		}
		GlobalOptions.Config = filepath.Join(home, ".volt", "config")
	}

	info, err := os.Stat(GlobalOptions.Config)
	if err == nil && !info.IsDir() {
		configFile, err := ioutil.ReadFile(GlobalOptions.Config)
		if err != nil {
			log.Printf("configFile.Get err   #%v ", err)
		}
		err = yaml.Unmarshal(configFile, &GlobalConfig)
		if err != nil {
			log.Fatalf("Unmarshal: %v", err)
		}
	}

	// Override from command line
	if GlobalOptions.Server != "" {
		GlobalConfig.Server = GlobalOptions.Server
	}
	if GlobalOptions.ApiVersion != "" {
		GlobalConfig.ApiVersion = GlobalOptions.ApiVersion
	}

	// If a k8s cert/key were not specified, then attempt to read it from
	// any $HOME/.kube/config if it exists
	if len(GlobalOptions.K8sConfig) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Unable to discover the user's home directory: %s\n", err)
			home = "~"
		}
		GlobalOptions.K8sConfig = filepath.Join(home, ".kube", "config")
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
				panic(err)
			}
			data, err = f.Process(data)
			if err != nil {
				panic(err)
			}
		}
		if result.OrderBy != "" {
			s, err := order.Parse(result.OrderBy)
			if err != nil {
				panic(err)
			}
			data, err = s.Process(data)
			if err != nil {
				panic(err)
			}
		}
		if result.OutputAs == OUTPUT_TABLE {
			tableFormat := format.Format(result.Format)
			tableFormat.Execute(os.Stdout, true, result.NameLimit, data)
		} else if result.OutputAs == OUTPUT_JSON {
			asJson, err := json.Marshal(&data)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s", asJson)
		} else if result.OutputAs == OUTPUT_YAML {
			asYaml, err := yaml.Marshal(&data)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s", asYaml)
		}
	}
}
