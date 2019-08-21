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
package order

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type Operation int

const (
	ASC Operation = iota
	DSC
)

type SortTerm struct {
	Op   Operation
	Name string
}

func (o Operation) String() string {
	switch o {
	default:
		fallthrough
	case ASC:
		return "ASC"
	case DSC:
		return "DSC"
	}
}

type Sorter []SortTerm

func split(term string) SortTerm {
	st := SortTerm{}
	if len(term) > 0 {
		switch term[0] {
		case '+':
			fallthrough
		case '>':
			st.Op = ASC
			st.Name = term[1:]
		case '-':
			fallthrough
		case '<':
			st.Op = DSC
			st.Name = term[1:]
		default:
			st.Op = ASC
			st.Name = term
		}
	} else {
		st.Op = ASC
		st.Name = term
	}
	return st
}

// Parse parses a comma separated list of filter terms
func Parse(spec string) (Sorter, error) {
	terms := strings.Split(spec, ",")
	s := make([]SortTerm, 0)
	for _, term := range terms {
		s = append(s, split(term))
	}

	return s, nil
}

func (s Sorter) Process(data interface{}) (interface{}, error) {
	slice := reflect.ValueOf(data)
	if slice.Kind() != reflect.Slice {
		return data, nil
	}

	sort.SliceStable(data, func(i, j int) bool {
		left := reflect.ValueOf(slice.Index(i).Interface())
		right := reflect.ValueOf(slice.Index(j).Interface())
		for _, term := range s {
			fleft := left.FieldByName(term.Name)
			fright := right.FieldByName(term.Name)
			switch fleft.Kind() {
			case reflect.Uint:
				fallthrough
			case reflect.Uint8:
				fallthrough
			case reflect.Uint16:
				fallthrough
			case reflect.Uint32:
				fallthrough
			case reflect.Uint64:
				ileft := fleft.Uint()
				iright := fright.Uint()
				switch term.Op {
				case ASC:
					if ileft < iright {
						return true
					} else if ileft > iright {
						return false
					}
				case DSC:
					if ileft > iright {
						return true
					} else if ileft < iright {
						return false
					}
				}
			case reflect.Int:
				fallthrough
			case reflect.Int8:
				fallthrough
			case reflect.Int16:
				fallthrough
			case reflect.Int32:
				fallthrough
			case reflect.Int64:
				ileft := fleft.Int()
				iright := fright.Int()
				switch term.Op {
				case ASC:
					if ileft < iright {
						return true
					} else if ileft > iright {
						return false
					}
				case DSC:
					if ileft > iright {
						return true
					} else if ileft < iright {
						return false
					}
				}
			default:
				sleft := fmt.Sprintf("%v", left.FieldByName(term.Name))
				sright := fmt.Sprintf("%v", right.FieldByName(term.Name))
				diff := strings.Compare(sleft, sright)
				if term.Op != DSC {
					if diff == -1 {
						return true
					} else if diff == 1 {
						return false
					}
				} else {
					if diff == 1 {
						return true
					} else if diff == -1 {
						return false
					}
				}
			}
		}
		return false
	})

	return data, nil
}
