# sdget
A tool for using DNS TXT records as a key/value table.

`sdget` is designed for use in scripts, and is mostly based on [RFC1464](https://tools.ietf.org/html/rfc1464) ("Using the Domain Name System To Store Arbitrary String Attributes") and [RFC6763](https://tools.ietf.org/html/rfc6763) ("DNS-Based Service Discovery").

## Usage examples

If `foo.example.com` has these TXT records:

```
foo.example.com. IN	TXT	"foo=bar"
foo.example.com. IN	TXT	"key=value"
```

they can be queried like this:
```bash
$ sdget foo.example.com key
value
```

Default values can be supplied, too:
```bash
$ sdget foo.example.com theanswer 42
42
```

Because `sdget` is designed for use in scripts and other automation, it's strict by default.  Missing TXT records, missing keys (without default values) or repeated keys are all treated as errors unless explicitly allowed using command line flags (NB: these flags are still not implemented).

## TXT format details
Each TXT string is treated as a simple key/value pair separated by a single `=`.  Everything after the first `=` is considered a value, which can contain any valid characters, including spaces or more `=` signs.  The key can contain any valid non-`=` characters.  Repeated keys are interpreted as lists.  Strings that aren't key/value pairs are simply ignored.

Here are some examples of valid key/value pairs:
```
foo=bar
org=Australian Digital Transformation Agency
Some Key=c29tZSBkYXRhCg==
list-key=1
list-key=2
list-key=3
empty value=
```

Note that [TXT records themselves have some size limitations](https://tools.ietf.org/html/rfc6763#section-6.1).
