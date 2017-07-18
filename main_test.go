package main

import (
	"errors"
	"reflect"
	"sort"
	"testing"
)

var sampleTxtRecords = []string{
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
	Options       *options
	TxtRecords    []string
	Key           string
	DefaultValues []string

	Result []string
	Err    error
}

func TestLookUpValues(t *testing.T) {
	for _, testPair := range []lookUpValuesTestPair{
		{makeDefaultOptions(), []string{}, "foo", []string{}, nil, errors.New("No values")},
		{makeDefaultOptions(), sampleTxtRecords, "foo", []string{}, []string{"bar"}, nil},
		{makeDefaultOptions(), sampleTxtRecords, "foo", []string{"default"}, []string{"bar"}, nil},
		{makeDefaultOptions(), sampleTxtRecords, "nosuchkey", []string{"default value"}, []string{"default value"}, nil},
		{makeDefaultOptions(), sampleTxtRecords, "empty", []string{}, []string{""}, nil},
		{makeDefaultOptions(), sampleTxtRecords, "spaces and multiple equals signs", []string{}, []string{"are no=problem at=all"}, nil},
		{makeDefaultOptions(), sampleTxtRecords, "nosuchkey", []string{}, nil, errors.New("No values")},
		{makeDefaultOptions(), sampleTxtRecords, "notkv", []string{}, nil, errors.New("No values")},
		{makeDefaultOptions(), sampleTxtRecords, "multival", []string{}, nil, errors.New("Too many values")},
	} {
		result, err := lookUpValues(testPair.Options, testPair.TxtRecords, testPair.Key, testPair.DefaultValues)

		if err != nil && testPair.Err == nil {
			t.Error("Unexpected error", err.Error(), "for", testPair)
		}

		if err == nil && testPair.Err != nil {
			t.Error("Expected error", testPair.Err.Error(), "not caught for", testPair)
		}

		sort.Strings(result)

		if !reflect.DeepEqual(result, testPair.Result) {
			t.Error("Expected", testPair.Result, "but got", result, "for", testPair)
		}
	}
}
