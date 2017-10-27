# sdget
A tool for using DNS TXT records as a key/value table.

`sdget` is designed for use in scripts, and is mostly based on [RFC1464](https://tools.ietf.org/html/rfc1464) ("Using the Domain Name System To Store Arbitrary String Attributes") and [RFC6763](https://tools.ietf.org/html/rfc6763) ("DNS-Based Service Discovery").

## Quick Start

If `foo.example.com` has these TXT records:

```
foo.example.com. IN	TXT	"foo=bar"
foo.example.com. IN	TXT	"key=value"
foo.example.com. IN	TXT	"things=item1"
foo.example.com. IN	TXT	"things=item2"
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

Because `sdget` is designed for use in scripts and other automation, it's strict by default.  If the domain has no TXT records, that's an error.  `sdget` also expects exactly one key to match, unless the `list` type is set:

```bash
$ sdget --type list foo.example.com things
item1
item2
```

The output format can be changed as well:

```bash
$ sdget --type list --format json foo.example.com things
["item1","item2"]
```

## Usage

```
usage: sdget [<flags>] <domain> <key> [<default>...]

Flags:
  -h, --help                   Show context-sensitive help (also try --help-long
                               and --help-man).
      --version                Show application version.
  -f, --format=plain           Output format (json, plain)
  -@, --nameserver=NAMESERVER  Nameserver address (ns.example.com:53, 127.0.0.1)
  -t, --type=single            Data value type (single, list)

Args:
  <domain>     Domain name to query for TXT records
  <key>        Key name to look up in domain
  [<default>]  Default value(s) to use if key is not found
```

Flag defaults can be set using environment variables of the form `SDGET_FLAGNAME`.  E.g.:

```bash
$ sdget foo.example.com key
value
$ export SDGET_FORMAT=json
$ sdget foo.example.com key
"value"
$ sdget --format plain foo.example.com key
value
$ unset SDGET_FORMAT
$ sdget foo.example.com key
value
```

### `--format`

* `json`: values encoded as JSON --- either a string or a list, depending on `--type`
* `plain`: values are output verbatim (default)

## TXT format details
Each TXT string is treated as a simple key/value pair separated by a single `=`.  Any `=` characters in the key name can be escaped using a backtick (`` ` ``), and everything after the first unescaped `=` is considered a value, which can contain any valid characters, including spaces or more `=` signs.  Keys are case-insensitive, and unescaped leading or trailing tabs and spaces are ignored.  Repeated keys are interpreted as lists.  Strings that aren't key/value pairs are simply ignored.

Here are some examples of valid key/value pairs:
```
foo=bar
org=Australian Digital Transformation Agency
Some Key=c29tZSBkYXRhCg==
list-key=1
list-key=2
list-key=3
empty value=
=empty key
 this key   =is stripped to just "this key"
` but this key` ` =keeps all its spaces
key`=with`=equals`=signs="value"
backticks``need=escaping in key names
but=not`in`values
```

The escaping rules only apply to the data in the TXT rules.  Keys supplied on the command line only need normal shell escaping rules.  See also the [original spec](https://tools.ietf.org/html/rfc1464#page-2).

Note that [TXT records themselves have some size limitations](https://tools.ietf.org/html/rfc6763#section-6.1).
