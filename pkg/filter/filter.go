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
package filter

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

type Operation int

const (
	UK Operation = iota
	EQ
	NE
	GT
	LT
	GE
	LE
	RE
)

func toOp(op string) Operation {
	switch op {
	case "=":
		return EQ
	case "!=":
		return NE
	case ">":
		return GT
	case "<":
		return LT
	case ">=":
		return GE
	case "<=":
		return LE
	case "~":
		return RE
	default:
		return UK
	}
}

type FilterTerm struct {
	Op    Operation
	Value string
	re    *regexp.Regexp
}

type Filter map[string]FilterTerm

var termRE = regexp.MustCompile("^\\s*([a-zA-Z_][.a-zA-Z0-9_]*)\\s*(~|<=|>=|<|>|!=|=)\\s*(.+)\\s*$")

// Parse parses a comma separated list of filter terms
func Parse(spec string) (Filter, error) {
	filter := make(map[string]FilterTerm)
	terms := strings.Split(spec, ",")
	var err error

	// Each term is in the form <key><op><value>
	for _, term := range terms {
		parts := termRE.FindAllStringSubmatch(term, -1)
		if parts == nil {
			return nil, fmt.Errorf("Unable to parse filter term '%s'", term)
		}
		ft := FilterTerm{
			Op:    toOp(parts[0][2]),
			Value: parts[0][3],
		}
		if ft.Op == RE {
			ft.re, err = regexp.Compile(ft.Value)
			if err != nil {
				return nil, fmt.Errorf("Unable to parse regexp filter value '%s'", ft.Value)
			}
		}
		filter[parts[0][1]] = ft
	}
	return filter, nil
}

func (f Filter) Process(data interface{}) (interface{}, error) {
	slice := reflect.ValueOf(data)
	if slice.Kind() != reflect.Slice {
		if f.Evaluate(data) {
			return data, nil
		}
		return nil, nil
	}

	var result []interface{}

	for i := 0; i < slice.Len(); i++ {
		if f.Evaluate(slice.Index(i).Interface()) {
			result = append(result, slice.Index(i).Interface())
		}
	}

	return result, nil
}

func (f Filter) Evaluate(item interface{}) bool {
	val := reflect.ValueOf(item)

	for k, v := range f {
		field := val.FieldByName(k)
		if !field.IsValid() {
			return false
		}

		switch v.Op {
		case RE:
			if !v.re.MatchString(fmt.Sprintf("%v", field)) {
				return false
			}
		case EQ:
			// This seems to work for most comparisons
			if fmt.Sprintf("%v", field) != v.Value {
				return false
			}
		default:
			// For unsupported operations, always pass
		}
	}
	return true
}
