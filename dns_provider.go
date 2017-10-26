package main

import (
	"io"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

type dnsProvider struct {
	options *options
	nameserver string
	domain string
}

func makeDnsProvider(options *options, nameserver string, domain string) (txtProvider, error) {
	if domain == "" {
		return nil, errors.New("non-empty domain name required")
	}
	nameserver, err := canonicalNameserver(options, nameserver)
	if err != nil {
		return nil, errors.Wrap(err, "error configuring DNS client")
	}
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	return &dnsProvider{
		options: options,
		nameserver: nameserver,
		domain: domain,
	}, nil
}

func (d *dnsProvider) getTxtRecords() ([]string, error) {
	query := new(dns.Msg)
	query.SetQuestion(d.domain, dns.TypeTXT)
	query.RecursionDesired = true

	client := new(dns.Client)
	response, _, err := client.Exchange(query, d.nameserver)
	if err == dns.ErrTruncated {
		client.Net = "tcp"
		response, _, err = client.Exchange(query, d.nameserver)
	}
	if err != nil {
		return nil, errors.Wrap(err, "error executing DNS query")
	}

	switch response.Rcode {
	case dns.RcodeSuccess:
		// okay

	case dns.RcodeNameError: // a.k.a. NXDOMAIN
		// TODO: add an option to allow ignoring this
		// This is the default for safety reasons
		return nil, errors.Errorf("no TXT records for domain %s", d.domain)

	default:
		return nil, errors.Errorf("error from remote DNS server: %s", dns.RcodeToString[response.Rcode])
	}

	var results []string
	for _, answer := range response.Answer {
		if txt, ok := answer.(*dns.TXT); ok {
			results = append(results, strings.Join(txt.Txt, ""))
		}
	}
	return results, nil
}

func canonicalNameserver(options *options, nameserver string) (string, error) {
	if nameserver == "" {
		if options.nameserver == "" {
			return configFromResolvConf(options)
		}
		nameserver = options.nameserver
	}
	return addNameserverPort(options, nameserver)
}

func configFromResolvConf(options *options) (string, error) {
	resolvconf, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return "", errors.Wrap(err, "error opening /etc/resolv.conf")
	}
	defer resolvconf.Close()
	return readResolvConf(options, resolvconf)
}

func readResolvConf(options *options, resolvconf io.Reader) (string, error) {
	config, err := dns.ClientConfigFromReader(resolvconf)
	if err != nil {
		return "", errors.Wrap(err, "error reading resolv.conf DNS configuration")
	}
	return config.Servers[0] + ":" + config.Port, nil
}

// Users can specify the nameserver host and port, or just the host, or neither.
// In the last case, we fall back to the system config.  Otherwise we need to make sure we have a port (53 as default).
func addNameserverPort(options *options, nameserver string) (string, error) {
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
	// [::1]
	//
	// IPv6's reuse of colons makes life more complicated.  IPv6 addresses without square brackets shouldn't appear in
	// URIs, but we need to handle them because they might come in through the --nameserver option.
	hasPortRegex := "^([^:]+|\\[.+\\]):[0-9]+$"
	hasPort, err := regexp.MatchString(hasPortRegex, nameserver)
	if err != nil {
		return "", errors.Wrap(err, "error reading nameserver address")
	}
	if !hasPort {
		unwrappedIPv6Regex := "^[0-9a-fA-F:]*:[0-9a-fA-F:]+$"
		needsWrapping, err := regexp.MatchString(unwrappedIPv6Regex, nameserver)
		if err != nil {
			return "", errors.Wrap(err, "error reading nameserver address")
		}
		if needsWrapping {
			nameserver = fmt.Sprintf("[%s]:53", nameserver)
		} else {
			nameserver = nameserver + ":53"
		}
	}
	return nameserver, nil
}
