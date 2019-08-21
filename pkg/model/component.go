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
package model

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"time"
)

type ComponentInstance struct {
	Id        string `json:"id"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Ready     string `json:"ready"`
	Status    string `json:"status"`
	Restarts  int    `json:"restarts"`
	Component string `json:"component"`
	Version   string `json:"version"`
	StartTime string `json:"starttime"`
	Age       string `json:"age"`
}

func (c *ComponentInstance) PopulateFrom(val corev1.Pod) {
	c.Id = val.ObjectMeta.Name
	c.Namespace = val.ObjectMeta.Namespace
	c.Name = val.ObjectMeta.Labels["app.kubernetes.io/name"]
	c.Component = val.ObjectMeta.Labels["app.kubernetes.io/component"]
	c.Version = val.ObjectMeta.Labels["app.kubernetes.io/version"]
	c.Status = string(val.Status.Phase)

	ready := 0
	var restarts int = 0
	for _, d := range val.Status.ContainerStatuses {
		if d.Ready {
			ready += 1
		}
		restarts += int(d.RestartCount)
	}
	c.Ready = fmt.Sprintf("%d/%d", ready, len(val.Status.ContainerStatuses))
	c.Restarts = restarts

	c.StartTime = val.Status.StartTime.Time.String()
	c.Age = time.Since(val.Status.StartTime.Time).Truncate(time.Second).String()
}
