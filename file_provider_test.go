package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestGetTxtRecords(t *testing.T) {
	sampleRecordsFile := strings.NewReader(`a=b

"quoted=line"
"quoted=line with \" escaped quote and \\ escaped backslash"
"escape=sequences\x21\x21\x21"
not a record
unquoted=\n\"record
` + "`\"unquoted=record")
	expected := []string{
		"a=b",
		"",
		"quoted=line",
		`quoted=line with " escaped quote and \ escaped backslash`,
		"escape=sequences!!!",
		"not a record",
		`unquoted=\n\"record`,
		"`\"unquoted=record",
	}
	options := makeDefaultOptions()
	provider, err := makeFileProvider(options, "localhost", "/test/path")
	if err != nil {
		t.Error("Error", err.Error())
	}

	records, err := provider.getTxtRecordsFromReader(sampleRecordsFile)
	if err != nil {
		t.Error("Error", err.Error())
	}

	if !reflect.DeepEqual(records, expected) {
		t.Error("Expected\n", expected, "\nbut got\n", records)
	}
}
