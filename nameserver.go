package main

import (
	"fmt"
	"io"
	"regexp"
	"os"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

// Users can specify the nameserver host and port, or just the host, or neither.
// In the last case, we fall back to the system config.  Otherwise we need to make sure we have a port (53 as default).

func configureNameserver(options *options) error {
	if options.nameserver == "" {
		resolvconf, err := os.Open("/etc/resolv.conf")
		if err != nil {
			return errors.Wrap(err, "error opening /etc/resolv.conf")
		}
		defer resolvconf.Close()
		return readResolvConf(options, resolvconf)
	}
	return addNameserverPort(options)
}

func readResolvConf(options *options, resolvconf io.Reader) error {
	config, err := dns.ClientConfigFromReader(resolvconf)
	if err != nil {
		return errors.Wrap(err, "error reading resolv.conf DNS configuration")
	}
	options.nameserver = config.Servers[0] + ":" + config.Port
	return nil
}

func addNameserverPort(options *options) error {
	// Addresses with ports:
	//
	// example.com:53
	// localhost:53
	// 127.0.0.1:53
	// [::1]:53
	//
	// Addresses without ports:
	//
	// example.com
	// localhost
	// 127.0.0.1
	// ::1
	//
	// (IPv6's reuse of colons makes life more complicated.)
	hasPortRegex := "^([^:]+|\\[.+\\]):[0-9]+$"
	hasPort, err := regexp.MatchString(hasPortRegex, options.nameserver)
	if err != nil {
		return errors.Wrap(err, "error reading nameserver address")
	}
	if !hasPort {
		unwrappedIPv6Regex := "^[0-9a-fA-F:]*:[0-9a-fA-F:]+$"
		needsWrapping, err := regexp.MatchString(unwrappedIPv6Regex, options.nameserver)
		if err != nil {
			return errors.Wrap(err, "error reading nameserver address")
		}
		if needsWrapping {
			options.nameserver = fmt.Sprintf("[%s]:53", options.nameserver)
		} else {
			options.nameserver = options.nameserver + ":53"
		}
	}
	return nil
}
