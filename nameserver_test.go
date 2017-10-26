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
}

func TestAddNameserverPort(t *testing.T) {
	for _, testPair := range []addNameserverPortTestPair{
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
		err := addNameserverPort(options)
		if err != nil {
			t.Error("Error", err.Error(), "for", testPair)
		}
		if options.nameserver != testPair.Result {
			t.Error("Expected", testPair.Result, "but got", options.nameserver)
		}
	}
}

func TestReadResolvConf(t *testing.T) {
	options := &options{}
	err := readResolvConf(options, resolvConf)
	if err != nil {
		t.Error("Error", err.Error())
	}
	if options.nameserver != defaultNameserver {
		t.Error("Expected", defaultNameserver, "but got", options.nameserver)
	}
}
