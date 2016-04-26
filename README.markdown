About
-----

This is a Go Sphinx client, using the Sphinx API.  It has is based on the
`github.com/yunge/sphinx` client, but has a number of important differences:

- Client is threadsafe to use in multiple goroutines simultaneously
- Pooled connections to Sphinx using the [github.com/fatih/pool](pool) library.
- Took out SphinxQL support
- Intended to work on Sphinx 2.0.8-release (r3831), so some things have been written
  with that goal in mind.  For example, doesn't do checking for old version when sending
  filter values, instead just uses new version (uint64 vs. uint32).
- Doesn't support GroupBy, GeoAnchor, SetSelect, or override options.

In addition, the approach to creating requests and reading / writing data to a
buffer uses a wrapped `bytes.Buffer` instead of a byte array directly - this
should be more efficient.

## Installation

`go get github.com/rewardStyle/sphinx`


## Testing

There is a vendored version of libsphinxclient in `fixture_data/libsphinx` which
can be built from source with `make`.  In `fixture_data`, there is a small helper
program which can be used to generate fixture data for requests and responses
for a given query.  Standard fixture data generated using this program is
included in `fixture_data/generated` and is used by the unit tests to verify correctness.

## Local Sphinx setup

Import "documents.sql" to "test" database in mysql;

Change the mysql password in sphinx.conf;

Copy the test.xml to default dir in sphinx.conf:
`cp test.xml /usr/local/sphinx/var/data`

Index the test data:
`indexer -c sphinx_lib_path/sphinx.conf --all --rotate`

Start sphinx searchd with "sphinx.conf":
`searchd -c sphinx_lib_path/sphinx.conf`

Then "cd" to sphinx_lib_path:

`go test`

## Examples
```Go
import (
  "github.com/rewardStyle/sphinx"
)

// TODO
```
More examples can be found in test files.

## LICENSE

(Changed from BSD -> GPLv2 so that we can vendor in libsphinxclient).

[GPL v2](http://www.gnu.org/licenses/old-licenses/gpl-2.0.html)
