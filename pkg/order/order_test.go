package order

import (
	"testing"
)

type SortTestStruct struct {
	Id    int
	One   string
	Two   string
	Three uint
	Four  int
}

var testSetOne = []SortTestStruct{
	{
		Id:    0,
		One:   "a",
		Two:   "x",
		Three: 10,
		Four:  1,
	},
	{
		Id:    1,
		One:   "a",
		Two:   "c",
		Three: 1,
		Four:  10,
	},
	{
		Id:    2,
		One:   "a",
		Two:   "b",
		Three: 2,
		Four:  1000,
	},
	{
		Id:    3,
		One:   "a",
		Two:   "a",
		Three: 3,
		Four:  100,
	},
	{
		Id:    4,
		One:   "b",
		Two:   "a",
		Three: 3,
		Four:  0,
	},
}

var testSetTwo = []SortTestStruct{
	{
		Id:    0,
		One:   "a",
		Two:   "x",
		Three: 10,
		Four:  10,
	},
	{
		Id:    1,
		One:   "a",
		Two:   "y",
		Three: 1,
		Four:  1,
	},
}

func Verify(v []SortTestStruct, order []int) bool {
	for i, item := range v {
		if item.Id != order[i] {
			return false
		}
	}
	return true
}

func TestSort(t *testing.T) {
	s, err := Parse("+One,-Two")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetOne)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	if !Verify(o.([]SortTestStruct), []int{0, 1, 2, 3, 4}) {
		t.Errorf("incorrect sort")
	}
}

func TestSortASC(t *testing.T) {
	s, err := Parse("+One,Two")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetTwo)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	if !Verify(o.([]SortTestStruct), []int{0, 1}) {
		t.Errorf("incorrect sort")
	}
}

func TestSortUintASC(t *testing.T) {
	s, err := Parse("Three,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetOne)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	if !Verify(o.([]SortTestStruct), []int{1, 2, 3, 4, 0}) {
		t.Errorf("incorrect sort")
	}
}

func TestSortUintDSC(t *testing.T) {
	s, err := Parse("-Three,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetOne)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	if !Verify(o.([]SortTestStruct), []int{0, 3, 4, 2, 1}) {
		t.Errorf("incorrect sort")
	}
}

func TestSortUintDSC2(t *testing.T) {
	s, err := Parse("-Three,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetTwo)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	if !Verify(o.([]SortTestStruct), []int{0, 1}) {
		t.Errorf("incorrect sort")
	}
}

func TestSortIntASC(t *testing.T) {
	s, err := Parse("Four,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetOne)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}
	if !Verify(o.([]SortTestStruct), []int{4, 0, 1, 3, 2}) {
		t.Errorf("incorrect sort")
	}
}

func TestSortIntDSC(t *testing.T) {
	s, err := Parse("-Four,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetOne)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}
	if !Verify(o.([]SortTestStruct), []int{2, 3, 1, 0, 4}) {
		t.Errorf("incorrect sort")
	}
}

func TestSortIntDSC2(t *testing.T) {
	s, err := Parse("-Four,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetTwo)
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}
	if !Verify(o.([]SortTestStruct), []int{0, 1}) {
		t.Errorf("incorrect sort")
	}
}

func TestOperString(t *testing.T) {
	if ASC.String() != "ASC" {
		t.Errorf("ASC to string failed")
	}
	if DSC.String() != "DSC" {
		t.Errorf("DSC to string failed")
	}
	var o Operation = 5 // Invalid
	if o.String() != "ASC" {
		t.Errorf("to string default failed")
	}
}

func TestSortSingle(t *testing.T) {
	s, err := Parse("-Four,One")
	if err != nil {
		t.Errorf("Unable to parse sort specification")
	}
	o, err := s.Process(testSetOne[0])
	if err != nil {
		t.Errorf("Sort failed: %s", err.Error())
	}

	if o == nil {
		t.Errorf("expected value, got nil")
	}

	r, ok := o.(SortTestStruct)
	if !ok {
		t.Errorf("Unexpected result type")
	}

	if r.Id != testSetOne[0].Id {
		t.Errorf("results don't match input")
	}
}
