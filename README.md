# tap13

[![GoDoc](https://godoc.org/github.com/mpontillo/tap13?status.svg)](https://godoc.org/github.com/mpontillo/tap13)

This pacakge implements a parser for the Test Anything Protocol (TAP)
version 13 specification.

The full protocol specification can be found at the following URL:

https://testanything.org/tap-version-13-specification.html

This package is intended to allow the results of a TAP test run to be
converted to a more useful format for storage and analysis.

# Usage as a Library

Given a `[]string` containing TAP v13 output, users can call the `Parse()`
method to return a `Results` struct. A `Stringer` interface is implemented
on the `Results` object in order to provide a summary of the test run.

# Usage as a command-line tool

A `tap13` command-line tool is provided. It will read the contents of
each file (assumed to contain TAP version 13 results) specified as an
argument, and output a summary of the test results.

This tool is primarily intended for testing the library itself.

# Bugs

The parser does not currently interpret or store YAML output, or comments
specified as directives on the same line as a test run.

Bugs are [tracked on the GitHub issues page](https://github.com/mpontillo/tap13/issues).

# Contributing

Pull requests are welcome! Any additional code must be backward compatible,
and covered appropriately by test cases.
