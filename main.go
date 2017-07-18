package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

type options struct{}

func makeDefaultOptions() *options {
	return &options{}
}

func printUsage() {
	fmt.Printf("Usage: %s domain key [default value] [default value] ...\n", os.Args[0])
	flag.PrintDefaults()
}

func getTxtRecords(options *options, domain string) ([]string, error) {
	if !strings.HasSuffix(domain, ".") {
		domain = domain + "."
	}
	// TODO: just pulling the config from /etc/resolv.conf works most of the time, but we really need options for the other times
	config, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		return nil, errors.Wrap(err, "error configuring DNS client")
	}
	// This DNS client library doesn't do retries or fallbacks to TCP, but it's good enough for a proof of concept.
	// Maybe replace all this with bindings to one of the more solid C libraries.
	client := new(dns.Client)
	query := new(dns.Msg)
	query.SetQuestion(domain, dns.TypeTXT)
	query.RecursionDesired = true

	response, _, err := client.Exchange(query, config.Servers[0]+":"+config.Port)
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

	if len(values) == 0 && len(defaultValues) == 0 {
		// TODO: make this check optional
		// This is just a default restriction
		return nil, errors.Errorf("no values found for key %s, and no default provided", key)
	}
	if len(values) > 1 {
		// TODO: make this check optional, too
		// This is just a default restriction
		return nil, errors.Errorf("%d values found for key %s, but only 1 was expected", len(values), key)
	}
	if len(values) == 0 {
		return defaultValues, nil
	}
	return values, nil
}

func main() {
	options := makeDefaultOptions()

	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		printUsage()
		os.Exit(2)
	}

	domain := args[0]
	key := args[1]
	defaultValues := args[2:]

	txtRecords, err := getTxtRecords(options, domain)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error doing DNS lookup:\n%+v\n", err.Error())
		os.Exit(3)
	}

	var values []string
	values, err = lookUpValues(options, txtRecords, key, defaultValues)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error looking up values for key \"%s\" in domain %s:\n%+v\n", key, domain, err.Error())
		os.Exit(4)
	}

	// Only line-by-line output supported for now
	for _, record := range values {
		fmt.Printf("%s\n", record)
	}
}
