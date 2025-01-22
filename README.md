# go-test-sift

When Go runs unit tests in parallel, their output becomes an interleaved mess, making debugging test failures difficult. `go-test-sift` restores order by regrouping all output for each test and its subtests into a clean, hierarchical structure, preserving the parent-child relationships.

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

### Test Filtering

The `-t` flag accepts a regular expression to filter which tests to process. It can be used:
- On its own to filter the default regrouped output
- Combined with `-l` or `-L` to filter which failures to show
- Combined with `-w` to filter which test outputs to write to files

Examples:
```sh
# Only show output for specific tests
go-test-sift -t "TestAuth.*" test.log

# Only summarise failures for specific tests
go-test-sift -t "TestAuth.*" -l test.log

# Only write specific test outputs to files
go-test-sift -t "TestAuth.*" -w test.log
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
