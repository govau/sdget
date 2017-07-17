package main

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

var sample_txt_records = []string{
	"foo=bar",
	"empty=",
	"multival=1",
	"multival=2",
	"multival=3",
	"something that's not a key/value pair",
	"notkv",
	"",
	"=",
	"spaces and multiple equals signs=are no=problem at=all",
}

type lookUpValuesTestPair struct {
	options *Options
	txt_records []string
	key string
	default_values []string

	result []string
	err error
}

var look_up_values_pairs = []lookUpValuesTestPair{
	{ MakeDefaultOptions(), []string{}, "foo", []string{}, nil, errors.New("No values") },
	{ MakeDefaultOptions(), sample_txt_records, "foo", []string{}, []string{"bar"}, nil },
	{ MakeDefaultOptions(), sample_txt_records, "foo", []string{"default"}, []string{"bar"}, nil },
	{ MakeDefaultOptions(), sample_txt_records, "nosuchkey", []string{"default value"}, []string{"default value"}, nil },
	{ MakeDefaultOptions(), sample_txt_records, "empty", []string{}, []string{""}, nil },
	{ MakeDefaultOptions(), sample_txt_records, "spaces and multiple equals signs", []string{}, []string{"are no=problem at=all"}, nil },
	{ MakeDefaultOptions(), sample_txt_records, "nosuchkey", []string{}, nil, errors.New("No values") },
	{ MakeDefaultOptions(), sample_txt_records, "notkv", []string{}, nil, errors.New("No values") },
	{ MakeDefaultOptions(), sample_txt_records, "multival", []string{}, nil, errors.New("Too many values") },
}

func TestLookUpValues(t *testing.T) {
	for _, test_pair := range look_up_values_pairs {
		result, err := LookUpValues(test_pair.options, test_pair.txt_records, test_pair.key, test_pair.default_values)

		if err != nil && test_pair.err == nil {
			t.Error("Unexpected error", err.Error(), "for", test_pair);
		}

		if err == nil && test_pair.err != nil {
			t.Error("Expected error", test_pair.err.Error(), "not caught for", test_pair);
		}

		sort.Strings(result)

		if !reflect.DeepEqual(result, test_pair.result) {
			t.Error("Expected", test_pair.result, "but got", result, "for", test_pair);
		}
	}
}
