package main

import (
	"strings"
	"testing"
)

var resolvConf = strings.NewReader("nameserver 192.168.7.1\n")
var defaultNameserver = "192.168.7.1:53"

type addNameserverPortTestPair struct {
	Input  string
	Result string
	Err    error
}

func TestAddNameserverPort(t *testing.T) {
	for _, testPair := range []addNameserverPortTestPair{
		{"example.com", "example.com:53", nil},
		{"example.com:1053", "example.com:1053", nil},
		{"::", "[::]:53", nil},
		{"::1", "[::1]:53", nil},
		{"1::", "[1::]:53", nil},
		{"[::]:1053", "[::]:1053", nil},
		{"2606:2800:220:1:248:1893:25c8:1946", "[2606:2800:220:1:248:1893:25c8:1946]:53", nil},
		{"[2606:2800:220:1:248:1893:25c8:1946]:1053", "[2606:2800:220:1:248:1893:25c8:1946]:1053", nil},
	} {
		options := makeDefaultOptions()
		nameserver, err := addNameserverPort(options, testPair.Input)
		if err != nil {
			t.Error("Error", err.Error(), "for", testPair)
		}
		if nameserver != testPair.Result {
			t.Error("Expected", testPair.Result, "but got", nameserver, "for", testPair)
		}
	}
}

func TestReadResolvConf(t *testing.T) {
	options := &options{}
	nameserver, err := readResolvConf(options, resolvConf)
	if err != nil {
		t.Error("Error", err.Error())
	}
	if nameserver != defaultNameserver {
		t.Error("Expected", defaultNameserver, "but got", nameserver)
	}
}
