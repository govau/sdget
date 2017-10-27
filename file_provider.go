package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/pkg/errors"
)

type fileProvider struct {
	options *options
	path    string
}

func makeFileProvider(options *options, hostname string, path string) (*fileProvider, error) {
	if hostname != "" && hostname != "localhost" {
		machineHostname, _ := os.Hostname()
		if hostname != machineHostname {
			return nil, fmt.Errorf("unsupported hostname in file URI: %s", hostname)
		}
	}
	return &fileProvider{
		options: options,
		path:    path,
	}, nil
}

func (f *fileProvider) getTxtRecords() ([]string, error) {
	file, err := os.Open(f.path)
	if err != nil {
		return nil, errors.Wrap(err, "error opening file for TXT records")
	}

	return f.getTxtRecordsFromReader(file)
}

func (f *fileProvider) getTxtRecordsFromReader(input io.Reader) ([]string, error) {
	var result []string
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		record := scanner.Text()
		unquoted, err := unquoteRecord(record)
		if err != nil {
			return nil, errors.Wrapf(err, "error unquoting TXT record \"%s\"", record)
		}
		result = append(result, unquoted)
	}

	err := scanner.Err()
	if err != nil {
		return nil, errors.Wrapf(err, "error reading file \"%s\" for TXT records", f.path)
	}
	return result, nil
}

func unquoteRecord(record string) (string, error) {
	if record == "" || record[0] != '"' {
		return record, nil
	}
	return strconv.Unquote(record)
}
