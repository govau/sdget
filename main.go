package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

type options struct {
	outputFormat string
	valueType    string
	nameserver   string
}

func makeDefaultOptions() *options {
	return &options{
		outputFormat: "plain",
		valueType:    "single",
	}
}

func output(options *options, sink io.Writer, values []string) error {
	if options.valueType == "single" && len(values) != 1 {
		return fmt.Errorf("expected 1 value but got %d (%v)", len(values), values)
	}
	switch options.outputFormat {
	case "json":
		var err error
		encoder := json.NewEncoder(sink)
		encoder.SetEscapeHTML(false)
		switch options.valueType {
		case "single":
			err = encoder.Encode(values[0])
		case "list":
			err = encoder.Encode(values)
		}
		if err != nil {
			return errors.Wrap(err, "error writing JSON")
		}
	case "plain":
		for _, record := range values {
			if _, err := fmt.Fprintln(sink, record); err != nil {
				return err
			}
		}
	}
	return nil
}

func configureNameserver(options *options, resolvconf io.Reader) error {
	// Users can specify the nameserver host and port, or just the host, or neither.
	// In the last case, we fall back to the system config.  Otherwise we need make sure we have a port (53 as default).
	if options.nameserver == "" {
		// FIXME: use the supplied resolvconf instead of the hardcoded path when this is merged:
		// https://github.com/miekg/dns/pull/532
		config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			return errors.Wrap(err, "error reading resolv.conf DNS configuration")
		}
		options.nameserver = config.Servers[0] + ":" + config.Port
	} else {
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
	}
	return nil
}

func getTxtRecords(options *options, domain string) ([]string, error) {
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}

	query := new(dns.Msg)
	query.SetQuestion(domain, dns.TypeTXT)
	query.RecursionDesired = true

	client := new(dns.Client)
	response, _, err := client.Exchange(query, options.nameserver)
	if err == dns.ErrTruncated {
		client.Net = "tcp"
		response, _, err = client.Exchange(query, options.nameserver)
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
		return nil, errors.Errorf("no TXT records for domain %s", domain)

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

func lookUpValues(options *options, txtRecords []string, key string, defaultValues []string) ([]string, error) {
	var values []string
	for _, record := range txtRecords {
		pieces := strings.SplitN(record, "=", 2)
		if len(pieces) > 1 && pieces[0] == key {
			values = append(values, pieces[1])
		}
	}
	if len(values) == 0 {
		values = defaultValues
	}

	if options.valueType == "single" {
		if len(values) == 0 {
			return nil, errors.Errorf("no values found for key %s, and no default provided", key)
		}
		if len(values) > 1 {
			return nil, errors.Errorf("%d values found for key %s, but only 1 was expected", len(values), key)
		}
	}

	return values, nil
}

func main() {
	options := makeDefaultOptions()
	kingpin.Version("0.4.0")
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Flag("format", "Output format (json, plain)").Short('f').Default("plain").Envar("SDGET_FORMAT").EnumVar(&options.outputFormat, "json", "plain")
	kingpin.Flag("nameserver", "Nameserver address (ns.example.com:53, 127.0.0.1)").Short('@').Envar("SDGET_NAMESERVER").StringVar(&options.nameserver)
	kingpin.Flag("type", "Data value type (single, list)").Short('t').Default("single").Envar("SDGET_TYPE").EnumVar(&options.valueType, "single", "list")
	domain := kingpin.Arg("domain", "Domain name to query for TXT records").Required().String()
	key := kingpin.Arg("key", "Key name to look up in domain").Required().String()
	defaultValues := kingpin.Arg("default", "Default value(s) to use if key is not found").Strings()
	kingpin.Parse()

	if *defaultValues == nil {
		defaultValues = &[]string{}
	}

	if options.valueType == "single" && len(*defaultValues) > 1 {
		fmt.Fprintf(os.Stderr, "Got %n default values, but the value type is \"single\".  (Did you mean to set --type list?)\n", len(*defaultValues))
		os.Exit(1)
	}

	if err := configureNameserver(options, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error configuring nameserver: %s\n", err.Error())
		os.Exit(2)
	}

	txtRecords, err := getTxtRecords(options, *domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error doing DNS lookup:\n%+v\n", err.Error())
		os.Exit(3)
	}

	var values []string
	values, err = lookUpValues(options, txtRecords, *key, *defaultValues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error looking up values for key \"%s\" in domain %s:\n%+v\n", *key, *domain, err.Error())
		os.Exit(4)
	}

	if err = output(options, os.Stdout, values); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output values: %s\n", err.Error())
		os.Exit(5)
	}
}
