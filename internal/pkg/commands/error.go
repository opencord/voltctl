/*
 * Portions copyright 2019-present Open Networking Foundation
 * Original copyright 2019-present Ciena Corporation
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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/grpc/status"
)

var (
	descRE      = regexp.MustCompile(`desc = "(.*)"`)
	try2RE      = regexp.MustCompile(`all SubConns are in TransientFailure, latest connection error: (.*)`)
	try3RE      = regexp.MustCompile(`all SubConns are in TransientFailure, (.*)`)
	NoReportErr = errors.New("no-report-please")
)

func ErrorToString(err error) string {
	if err == nil {
		return ""
	}

	if st, ok := status.FromError(err); ok {
		msg := st.Message()
		if match := descRE.FindAllStringSubmatch(msg, 1); match != nil {
			msg = match[0][1]
		} else if match = try2RE.FindAllStringSubmatch(msg, 1); match != nil {
			msg = match[0][1]
		} else if match = try3RE.FindAllStringSubmatch(msg, 1); match != nil {
			msg = match[0][1]
		}

		return fmt.Sprintf("%s: %s", strings.ToUpper(st.Code().String()), msg)
	}
	return err.Error()
}
