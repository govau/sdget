package main

import (
	"bytes"
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
	"`=",
	"`",
	"spaces and multiple equals signs=are no=problem at=all",
	"\t with tabs\tand spaces \t=  whitespace value\t ",
	"CamelCase=CamelCaseValue",
	"ALLCAPS=ALLCAPS VALUE",
	"` key`=with escapes`\t \t=`escaped`value",
}

var defaultOptions = makeDefaultOptions()
var plainListOptions = &options{
	outputFormat: "plain",
	valueType:    "list",
}
var jsonSingleOptions = &options{
	outputFormat: "json",
	valueType:    "single",
}
var jsonListOptions = &options{
	outputFormat: "json",
	valueType:    "list",
}
var zeroListOptions = &options{
	outputFormat: "zero",
	valueType:    "list",
}

type outputTestPair struct {
	Options *options
	Values  []string
	Result  string
	Err     error
}

func TestOutput(t *testing.T) {
	for _, testPair := range []outputTestPair{
		{defaultOptions, []string{}, "", errors.New("No values")},
		{defaultOptions, []string{"foo"}, "foo\n", nil},
		{defaultOptions, []string{"foo", "bar"}, "", errors.New("Too many values")},
		{plainListOptions, []string{}, "", nil},
		{plainListOptions, []string{"foo"}, "foo\n", nil},
		{plainListOptions, []string{"foo", "bar"}, "foo\nbar\n", nil},
		{jsonSingleOptions, []string{}, "", errors.New("No values")},
		{jsonSingleOptions, []string{"foo"}, "\"foo\"\n", nil},
		{jsonSingleOptions, []string{"foo", "bar"}, "", errors.New("Too many values")},
		{jsonSingleOptions, []string{"\"foo\""}, "\"\\\"foo\\\"\"\n", nil},
		{jsonSingleOptions, []string{"{}"}, "\"{}\"\n", nil},
		{jsonListOptions, []string{}, "[]\n", nil},
		{jsonListOptions, []string{"foo"}, "[\"foo\"]\n", nil},
		{jsonListOptions, []string{"foo", "bar"}, "[\"foo\",\"bar\"]\n", nil},
		{jsonListOptions, []string{"[]"}, "[\"[]\"]\n", nil},
		{zeroListOptions, []string{"a", "b", "c"}, "a\000b\000c\000", nil},
		{zeroListOptions, []string{`"foo and bar"`}, "\"foo and bar\"\000", nil},
	} {
		var outBuffer bytes.Buffer
		err := output(testPair.Options, &outBuffer, testPair.Values)
		result := outBuffer.String()

		if err != nil && testPair.Err == nil {
			t.Error("Unexpected error", err.Error(), "for", testPair)
		}

		if err == nil && testPair.Err != nil {
			t.Error("Expected error", testPair.Err.Error(), "not caught for", testPair)
		}

		if result != testPair.Result {
			t.Error("Expected", testPair.Result, "but got", result, "for", testPair)
		}
	}
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
		{defaultOptions, []string{}, "foo", []string{}, nil, errors.New("No values")},
		{defaultOptions, sampleTxtRecords, "foo", []string{}, []string{"bar"}, nil},
		{defaultOptions, sampleTxtRecords, "FoO", []string{}, []string{"bar"}, nil},
		{defaultOptions, sampleTxtRecords, "foo", []string{"default"}, []string{"bar"}, nil},
		{defaultOptions, sampleTxtRecords, "nosuchkey", []string{"default value"}, []string{"default value"}, nil},
		{defaultOptions, sampleTxtRecords, "empty", []string{}, []string{""}, nil},
		{defaultOptions, sampleTxtRecords, "spaces and multiple equals signs", []string{}, []string{"are no=problem at=all"}, nil},
		{defaultOptions, sampleTxtRecords, "nosuchkey", []string{}, nil, errors.New("No values")},
		{defaultOptions, sampleTxtRecords, "notkv", []string{}, nil, errors.New("No values")},
		{defaultOptions, sampleTxtRecords, "multival", []string{}, nil, errors.New("Too many values")},
		{defaultOptions, sampleTxtRecords, "cAmElCaSe", []string{}, []string{"CamelCaseValue"}, nil},
		{defaultOptions, sampleTxtRecords, "Allcaps", []string{}, []string{"ALLCAPS VALUE"}, nil},
		{defaultOptions, sampleTxtRecords, "with tabs\tand spaces", []string{}, []string{"  whitespace value\t "}, nil},
		{defaultOptions, sampleTxtRecords, " key=with escapes\t", []string{}, []string{"`escaped`value"}, nil},
		{plainListOptions, sampleTxtRecords, "multival", []string{}, []string{"1", "2", "3"}, nil},
		{plainListOptions, sampleTxtRecords, "foo", []string{}, []string{"bar"}, nil},
		{plainListOptions, sampleTxtRecords, "nosuchkey", []string{}, []string{}, nil},
		{plainListOptions, sampleTxtRecords, "nosuchkey", []string{"1", "2"}, []string{"1", "2"}, nil},
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
