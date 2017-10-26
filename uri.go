package main

// Parser for URIs
// The one in net/url doesn't work properly for general URIs

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/pkg/errors"
)

type parsedURI struct {
	scheme    string
	authority string
	path      string
	query     string
	fragment  string
}

// https://tools.ietf.org/html/rfc3986
var parser = regexp.MustCompile(`^(?P<scheme>[a-z][a-z0-9+.-]*):(?://(?P<authority>[^/?#]*))?(?P<path>[^#?]*)?(?P<query>\?[^#]*)?(?P<fragment>#.*)?$`)

func parseURI(uri string) (*parsedURI, error) {
	matches := parser.FindStringSubmatch(uri)
	if matches == nil {
		return nil, fmt.Errorf("failed to parse \"%s\" as URI", uri)
	}
	result := &parsedURI{
		scheme:    matches[1],
		authority: matches[2],
		path:      matches[3],
		query:     matches[4],
		fragment:  matches[5],
	}
	var err error
	result.path, err = url.PathUnescape(result.path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unencode path component \"%s\" of URI \"%s\"", result.path, uri)
	}
	result.query, err = url.QueryUnescape(result.query)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unencode query component \"%s\" of URI \"%s\"", result.query, uri)
	}
	return result, nil
}
