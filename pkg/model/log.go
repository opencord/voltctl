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

type LogLevel struct {
	ComponentName string
	PackageName   string
	Level         string
}

func (logLevel *LogLevel) PopulateFrom(componentName, packageName, level string) {
	logLevel.ComponentName = componentName
	logLevel.PackageName = packageName
	logLevel.Level = level
}

type LogFeature struct {
	ComponentName string
	Status        string
}

func (logFeature *LogFeature) PopulateFrom(componentName, status string) {
	logFeature.ComponentName = componentName
	logFeature.Status = status
}
