package main

import (
	"errors"
	"reflect"
	"testing"
)

type parseURITestPair struct {
	Input  string
	Result *parsedURI
	Err    error
}

func TestParseURI(t *testing.T) {
	for _, testPair := range []parseURITestPair{
		{"", nil, errors.New("Not a URI")},
		{"nope", nil, errors.New("Not a URI")},
		// Examples from https://tools.ietf.org/html/rfc4501
		{"dns:www.example.org.?clAsS=IN;tYpE=A", &parsedURI{"dns", "", "www.example.org.", "?clAsS=IN;tYpE=A", ""}, nil},
		{"dns:www.example.org", &parsedURI{"dns", "", "www.example.org", "", ""}, nil},
		{"dns://192.168.1.1/ftp.example.org?type=A", &parsedURI{"dns", "192.168.1.1", "/ftp.example.org", "?type=A", ""}, nil},
		{"dns:world%20wide%20web.example%5c.domain.org?TYPE=TXT", &parsedURI{"dns", "", "world wide web.example\\.domain.org", "?TYPE=TXT", ""}, nil},
	} {
		result, err := parseURI(testPair.Input)

		if err != nil && testPair.Err == nil {
			t.Error("Unexpected error", err.Error(), "for", testPair)
		}

		if err == nil && testPair.Err != nil {
			t.Error("Expected error", testPair.Err.Error(), "not caught for", testPair)
		}

		if !reflect.DeepEqual(result, testPair.Result) {
			t.Error("Expected", testPair.Result, "but got", result, "for", testPair)
		}
	}
}
