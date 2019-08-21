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
package version

// Default build-time variable.
// These values can (should) be overridden via ldflags when built with
// `make`
var (
	Version   = "unknown-version"
	GoVersion = "unknown-goversion"
	VcsRef    = "unknown-vcsref"
	VcsDirty  = "unknown-vcsdirty"
	BuildTime = "unknown-buildtime"
	Os        = "unknown-os"
	Arch      = "unknown-arch"
)
