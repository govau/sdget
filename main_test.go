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
	"spaces and multiple equals signs=are no=problem at=all",
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

var resolvConf = bytes.NewBufferString("nameserver 127.0.0.1\n")

type configureNameserverTestPair struct {
	Input  string
	Result string
}

func TestConfigureNameserver(t *testing.T) {
	for _, testPair := range []configureNameserverTestPair{
		// FIXME: uncomment when FIXME in target function is fixed
		// {"", "127.0.0.1:53"},
		{"example.com", "example.com:53"},
		{"example.com:1053", "example.com:1053"},
		{"::", "[::]:53"},
		{"::1", "[::1]:53"},
		{"1::", "[1::]:53"},
		{"[::]:1053", "[::]:1053"},
		{"2606:2800:220:1:248:1893:25c8:1946", "[2606:2800:220:1:248:1893:25c8:1946]:53"},
		{"[2606:2800:220:1:248:1893:25c8:1946]:1053", "[2606:2800:220:1:248:1893:25c8:1946]:1053"},
	} {
		options := makeDefaultOptions()
		options.nameserver = testPair.Input
		err := configureNameserver(options, resolvConf)
		if err != nil {
			t.Error("Error", err.Error(), "for", testPair)
		}
		if options.nameserver != testPair.Result {
			t.Error("Expected", testPair.Result, "but got", options.nameserver)
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
		{defaultOptions, sampleTxtRecords, "foo", []string{"default"}, []string{"bar"}, nil},
		{defaultOptions, sampleTxtRecords, "nosuchkey", []string{"default value"}, []string{"default value"}, nil},
		{defaultOptions, sampleTxtRecords, "empty", []string{}, []string{""}, nil},
		{defaultOptions, sampleTxtRecords, "spaces and multiple equals signs", []string{}, []string{"are no=problem at=all"}, nil},
		{defaultOptions, sampleTxtRecords, "nosuchkey", []string{}, nil, errors.New("No values")},
		{defaultOptions, sampleTxtRecords, "notkv", []string{}, nil, errors.New("No values")},
		{defaultOptions, sampleTxtRecords, "multival", []string{}, nil, errors.New("Too many values")},
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
