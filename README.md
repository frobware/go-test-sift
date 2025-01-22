# go-test-sift

A tool for processing Go test output that reconstructs and regroups parallel test logs into their logical structure.

## Why?

When Go runs tests in parallel, the output from different tests gets interleaved in the log. `go-test-sift` automatically regroups this output so all lines from each test are collected together, maintaining the proper parent/child test hierarchy. This makes test logs much easier to read and debug.

## Installation

```sh
go install github.com/frobware/go-test-sift@latest
```

## Usage

The simplest case is to pipe Go test output directly to `go-test-sift`:

```sh
go test ./... -v | go-test-sift
```

This will regroup all the interleaved parallel test output into a clean, hierarchical format where each test's output is kept together.

You can also process existing log files or URLs:

```sh
go-test-sift test.log
go-test-sift https://path/to/test.log
```

### Additional Options

To just see test failures:
```sh
go-test-sift -l test.log     # Shows failed test names
go-test-sift -L test.log     # Shows failed tests with their output
```

To save regrouped output to files:
```sh
go-test-sift -w test.log     # Creates directory structure by test name
```

To filter specific tests:
```sh
go-test-sift -t "TestAuth.*" test.log  # Only process tests matching pattern
```

### Synopsis

```sh
go-test-sift [options] [file|url ...]
  -F	Force directory creation even if directories exist
  -L	Print summary of failures and include the full output for each failure
  -d	Enable debug output
  -l	Print summary of failures (list test names with failures)
  -o string
        Base directory to write output files (default current directory) (default ".")
  -t string
        Regular expression to filter test names for summary output (default ".*")
  -w	Write each test's output to individual files
```
