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
package format

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"time"
)

// formats a Timestamp proto as a RFC3339 date string
func formatTimestamp(tsproto *timestamppb.Timestamp) (string, error) {
	if tsproto == nil {
		return "", nil
	}
	ts, err := ptypes.Timestamp(tsproto)
	if err != nil {
		return "", err
	}
	return ts.Truncate(time.Second).Format(time.RFC3339), nil
}

func formatRfc3339(in interface{}) (string, error) {
	if in == nil {
		return "", nil
	}

	switch v := in.(type) {
	case time.Time:
		return v.Truncate(time.Second).Format(time.RFC3339), nil
	case timestamppb.Timestamp:
		ts, err := ptypes.Timestamp(&v)
		if err != nil {
			return "", err
		}
		return ts.Truncate(time.Second).Format(time.RFC3339), nil
	default:
		return "", fmt.Errorf("invalid interface type encounterd while formatting in rfc3339 format")
	}
}

// Computes the age of a timestamp and returns it in HMS format
func formatSince(tsproto *timestamppb.Timestamp) (string, error) {
	if tsproto == nil {
		return "", nil
	}
	ts, err := ptypes.Timestamp(tsproto)
	if err != nil {
		return "", err
	}
	return time.Since(ts).Truncate(time.Second).String(), nil
}
