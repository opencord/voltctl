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
	"encoding/json"
	"github.com/ciena/voltctl/internal/pkg/cli/version"
	"github.com/ciena/voltctl/pkg/format"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"strings"
)

type VersionDetails struct {
	Version   string `json:"version"`
	GoVersion string `json:"goversion"`
	VcsRef    string `json:"gitcommit"`
	VcsDirty  string `json:"gitdirty"`
	BuildTime string `json:"buildtime"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

type VersionOutput struct {
	Client  VersionDetails `json:"client"`
	Cluster VersionDetails `json:"cluster"`
}

type VersionOpts struct {
	OutputOptions
	ClientOnly bool `long:"clientonly" description:"Display only client version information"`
}

var versionOpts = VersionOpts{}

var versionInfo = VersionOutput{
	Client: VersionDetails{
		Version:   version.Version,
		GoVersion: version.GoVersion,
		VcsRef:    version.VcsRef,
		VcsDirty:  version.VcsDirty,
		Os:        version.Os,
		Arch:      version.Arch,
		BuildTime: version.BuildTime,
	},
	Cluster: VersionDetails{
		Version:   "unknown-version",
		GoVersion: "unknown-goversion",
		VcsRef:    "unknown-vcsref",
		VcsDirty:  "unknown-vcsdirty",
		Os:        "unknown-os",
		Arch:      "unknown-arch",
		BuildTime: "unknown-buildtime",
	},
}

func RegisterVersionCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("version", "display version", "Display client and server version", &versionOpts)
	if err != nil {
		panic(err)
	}
}

const ClientOnlyFormat = `Client:
 Version        {{.Version}}
 Go version:    {{.GoVersion}}
 Vcs reference: {{.VcsRef}}
 Vcs dirty:     {{.VcsDirty}}
 Built:         {{.BuildTime}}
 OS/Arch:       {{.Os}}/{{.Arch}}
`

const DefaultFormat = `Client:
 Version        {{.Client.Version}}
 Go version:    {{.Client.GoVersion}}
 Vcs reference: {{.Client.VcsRef}}
 Vcs dirty:     {{.Client.VcsDirty}}
 Built:         {{.Client.BuildTime}}
 OS/Arch:       {{.Client.Os}}/{{.Client.Arch}}

Cluster:
 Version        {{.Cluster.Version}}
 Go version:    {{.Cluster.GoVersion}}
 Vcs feference: {{.Cluster.VcsRef}}
 Vcs dirty:     {{.Cluster.VcsDirty}}
 Built:         {{.Cluster.BuildTime}}
 OS/Arch:       {{.Cluster.Os}}/{{.Cluster.Arch}}
`

func (options *VersionOpts) clientOnlyVersion(args []string) error {
	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = ClientOnlyFormat
	}
	if options.Quiet {
		outputFormat = "{{.Version}}"
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      versionInfo.Client,
	}

	GenerateOutput(&result)
	return nil
}

func (options *VersionOpts) Execute(args []string) error {
	if options.ClientOnly {
		return options.clientOnlyVersion(args)
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	descriptor, method, err := GetMethod("version")
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
	version, err := d.TryGetFieldByName("version")
	if err != nil {
		return err
	}

	info := make(map[string]interface{})
	err = json.Unmarshal([]byte(version.(string)), &info)
	if err != nil {
		versionInfo.Cluster.Version = strings.ReplaceAll(version.(string), "\n", "")
	} else {
		var ok bool
		if _, ok = info["version"]; ok {
			versionInfo.Cluster.Version = info["version"].(string)
		} else {
			versionInfo.Cluster.Version = "unknown-version"
		}

		if _, ok = info["goversion"]; ok {
			versionInfo.Cluster.GoVersion = info["goversion"].(string)
		} else {
			versionInfo.Cluster.GoVersion = "unknown-goversion"
		}

		if _, ok = info["vcsref"]; ok {
			versionInfo.Cluster.VcsRef = info["vcsref"].(string)
		} else {
			versionInfo.Cluster.VcsRef = "unknown-vcsref"
		}

		if _, ok = info["vcsdirty"]; ok {
			versionInfo.Cluster.VcsDirty = info["vcsdirty"].(string)
		} else {
			versionInfo.Cluster.VcsDirty = "unknown-vcsdirty"
		}

		if _, ok = info["buildtime"]; ok {
			versionInfo.Cluster.BuildTime = info["buildtime"].(string)
		} else {
			versionInfo.Cluster.BuildTime = "unknown-buildtime"
		}

		if _, ok = info["os"]; ok {
			versionInfo.Cluster.Os = info["os"].(string)
		} else {
			versionInfo.Cluster.Os = "unknown-os"
		}

		if _, ok = info["arch"]; ok {
			versionInfo.Cluster.Arch = info["arch"].(string)
		} else {
			versionInfo.Cluster.Arch = "unknown-arch"
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DefaultFormat
	}
	if options.Quiet {
		outputFormat = "{{.Client.Version}}"
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      versionInfo,
	}

	GenerateOutput(&result)
	return nil
}
