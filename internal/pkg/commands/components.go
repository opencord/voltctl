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
	"fmt"
	"github.com/ciena/voltctl/pkg/format"
	"github.com/ciena/voltctl/pkg/model"
	flags "github.com/jessevdk/go-flags"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

const (
	DEFAULT_COMPONENT_FORMAT   = "table{{.Namespace}}\t{{.Id}}\t{{.Name}}\t{{.Component}}\t{{.Version}}\t{{.Ready}}\t{{.Restarts}}\t{{.Status}}\t{{.Age}}"
	COMPONENT_LIST_KUBECTL_CMD = "kubectl get --all-namespaces pod -l app.kubernetes.io/part-of=voltha -L app.kubernetes.io/name,app.kubernetes.io/component,app.kubernetes.io/version"
)

type ComponentList struct {
	ListOutputOptions
	Kubectl bool `long:"kubectl" short:"k" description:"display the kubectl command to execute"`
}

type ComponentOpts struct {
	List ComponentList `command:"list"`
}

var componentOpts = ComponentOpts{}

func RegisterComponentCommands(parser *flags.Parser) {
	parser.AddCommand("component", "component instance commands", "Commands to query and manipulate VOLTHA component instances", &componentOpts)
}

func (options *ComponentList) Execute(args []string) error {

	ProcessGlobalOptions()

	// If they requested the source to the kubectl command that
	// can give the same information, then print it and return
	if options.Kubectl {
		fmt.Println(COMPONENT_LIST_KUBECTL_CMD)
		return nil
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", GlobalOptions.K8sConfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/part-of=voltha",
	})
	if err != nil {
		panic(err.Error())
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_COMPONENT_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Metadata.Name}}"
	}
	data := make([]model.ComponentInstance, len(pods.Items))
	for i, item := range pods.Items {
		data[i].PopulateFrom(item)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		Filter:    options.Filter,
		OrderBy:   options.OrderBy,
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}

	GenerateOutput(&result)
	return nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
