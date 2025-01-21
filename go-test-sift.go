/*
go-test-sift: A utility to process Go unit test logs and provide summary information or extract individual test outputs.

Feature: Reconstructing Parallel Unit Test Output

This program attempts to group interleaved log lines into the correct
test context, even for parallel or nested subtests. While Go's
parallel test execution can sometimes produce interleaved logs that
make attribution challenging (where subtests or parent tests are
running in parallel), the program does its best to maintain logical
groupings. Also note, output not associated with `t.Log`, `t.Error`,
or similar Go testing functions) might not always be attributed
correctly. Debug output (`-d`) can help troubleshoot such scenarios.

Options:

  -l  Print a summary of failed tests, including their names,
      statuses, and durations.

  -L  Print a summary of failed tests (like -l) but also include the
      full output for each failed test. Useful for debugging failures
      directly in the console.

  -d  Enable debug output. Provides detailed information about how the
      program processes the input, including line-by-line analysis and
      test context switching.

  -s  Serialise all test output to stdout. This mode prints each
      test's output alongside its summary, without creating any output
      files.

  -w  Write each test's output to individual files, organised by test
      name. Files are created in a directory structure that matches
      the test hierarchy.

  -F  Force directory creation when used with -w. If directories
      already exist, they are overwritten instead of causing the
      program to exit.

  -o  Base directory for writing output files (default: current
      directory). Used in conjunction with -w to specify where the
      test outputs should be saved.

  -t  Regular expression to filter test names for summary output. By
      default, all tests are included (".*"). You can use this option
      to restrict the output to specific tests.
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TestSummary struct {
	Status string
	Name   string
	Time   string
	Level  int
}

func main() {
	listFailures := flag.Bool("l", false, "Print summary of failures (list test names with failures)")
	listFailuresWithOutput := flag.Bool("L", false, "Print summary of failures and include the full output for each failure")
	forceFlag := flag.Bool("F", false, "Force directory creation even if directories exist")
	debugFlag := flag.Bool("d", false, "Enable debug output")
	serialiseFlag := flag.Bool("s", false, "Serialise all output to stdout without summarising or writing directories")
	writeFiles := flag.Bool("w", false, "Write each test's output to individual files")
	outputDir := flag.String("o", ".", "Base directory to write output files (default current directory)")
	testPattern := flag.String("t", ".*", "Regular expression to filter test names for summary output")

	flag.Parse()

	reTest, err := regexp.Compile(*testPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid regular expression for -t: %v\n", err)
		os.Exit(1)
	}

	args := flag.Args()

	runRegex := regexp.MustCompile(`^=== RUN\s+(.*)`)
	nameRegex := regexp.MustCompile(`^=== NAME\s+(.*)`)
	contRegex := regexp.MustCompile(`^=== CONT\s+(.*)`)

	testBuffers := make(map[string][]string)
	var summaryRecords []TestSummary

	processReader := func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		var currentTest string
		var lastName string
		started := false
		inSummary := false
		lineNumber := 0
		stopParsing := false
		for scanner.Scan() {
			if stopParsing {
				break
			}

			lineNumber++
			line := scanner.Text()

			if *debugFlag {
				fmt.Printf("[DEBUG] Reading line %d: %s\n", lineNumber, line)
			}

			// Skip lines until the first '=== RUN' marker.
			if !started {
				if runRegex.MatchString(line) {
					started = true
					if *debugFlag {
						fmt.Println("[DEBUG] Found first '=== RUN' marker, starting parsing.")
					}
				} else {
					continue
				}
			}

			// Switch to summary mode when encountering a summary line.
			if !inSummary && strings.HasPrefix(strings.TrimSpace(line), "---") {
				inSummary = true
				if *debugFlag {
					fmt.Printf("[DEBUG] Encountered first summary line at line %d, switching to summary mode.\n", lineNumber)
				}
			}

			if inSummary {
				trimmed := strings.TrimSpace(line)
				if trimmed == "FAIL" {
					stopParsing = true
					if *debugFlag {
						fmt.Printf("[DEBUG] Encountered FAIL terminator at line %d; parsing now stopped\n", lineNumber)
					}
					continue
				}
				if strings.HasPrefix(trimmed, "---") {
					fields := strings.Fields(trimmed)
					// Expecting format: --- STATUS: TestName (Time).
					if len(fields) >= 3 {
						statusWithColon := fields[1]
						status := strings.TrimSuffix(statusWithColon, ":")
						testName := fields[2]
						timeStr := ""
						if len(fields) >= 4 {
							timeStr = strings.Trim(fields[3], "()")
						}
						indent := len(line) - len(strings.TrimLeft(line, " "))
						level := indent / 4

						newSummary := TestSummary{
							Status: status,
							Name:   testName,
							Time:   timeStr,
							Level:  level,
						}
						summaryRecords = append(summaryRecords, newSummary)
						if *debugFlag {
							fmt.Printf("[DEBUG] Added summary record at line %d: %+v\n", lineNumber, newSummary)
						}
					}
				}
				continue
			}

			switch {
			case strings.HasPrefix(strings.TrimSpace(line), "=== PAUSE"):
				if *debugFlag {
					fmt.Printf("[DEBUG] Encountered PAUSE for test context at line %d\n", lineNumber)
				}
			case contRegex.MatchString(line):
				match := contRegex.FindStringSubmatch(line)
				currentTest = match[1]
				lastName = currentTest
				if *debugFlag {
					fmt.Printf("[DEBUG] Resuming test context: %s at line %d\n", currentTest, lineNumber)
				}
			case nameRegex.MatchString(line):
				match := nameRegex.FindStringSubmatch(line)
				newName := match[1]

				if lastName != "" && strings.HasPrefix(lastName, newName+"/") {
					currentTest = lastName
					if *debugFlag {
						fmt.Printf("[DEBUG] Continuing context for subtest: %s at line %d\n", currentTest, lineNumber)
					}
				} else {
					currentTest = newName
					lastName = newName
					if *debugFlag {
						fmt.Printf("[DEBUG] Switching context to test: %s at line %d\n", currentTest, lineNumber)
					}
				}
			case runRegex.MatchString(line):
				match := runRegex.FindStringSubmatch(line)
				currentTest = match[1]
				if _, exists := testBuffers[currentTest]; !exists {
					testBuffers[currentTest] = []string{}
					if *debugFlag {
						fmt.Printf("[DEBUG] New test started at line %d: %s\n", lineNumber, currentTest)
					}
				}
			default:
				if currentTest != "" {
					testBuffers[currentTest] = append(testBuffers[currentTest], line)
					if *debugFlag {
						fmt.Printf("[DEBUG] Collected meaningful line %d for test %s\n", lineNumber, currentTest)
					}
				} else {
					fmt.Println(line)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		}
	}

	if len(args) == 0 {
		if *debugFlag {
			fmt.Println("[DEBUG] No arguments provided, reading from stdin.")
		}
		processReader(os.Stdin)
	} else {
		for _, input := range args {
			var reader io.Reader
			if u, err := url.ParseRequestURI(input); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
				if *debugFlag {
					fmt.Printf("[DEBUG] Fetching URL: %s\n", input)
				}
				resp, err := http.Get(input)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error fetching URL %s: %v\n", input, err)
					continue
				}
				defer resp.Body.Close()
				reader = resp.Body
			} else {
				if *debugFlag {
					fmt.Printf("[DEBUG] Opening file: %s\n", input)
				}
				file, err := os.Open(input)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", input, err)
					continue
				}
				defer file.Close()
				reader = file
			}
			processReader(reader)
		}
	}

	if *serialiseFlag {
		for _, summary := range summaryRecords {
			if !reTest.MatchString(summary.Name) {
				continue
			}
			indent := strings.Repeat("    ", summary.Level)
			fmt.Printf("%s--- %s: %s (%s)\n", indent, summary.Status, summary.Name, summary.Time)
			if lines, exists := testBuffers[summary.Name]; exists {
				for _, line := range lines {
					fmt.Printf("%s    %s\n", indent, line)
				}
			}
		}
		return
	}

	if *listFailures || *listFailuresWithOutput {
		fmt.Println("Failed Tests:")
		for _, summary := range summaryRecords {
			if summary.Status == "FAIL" && reTest.MatchString(summary.Name) {
				indent := strings.Repeat("    ", summary.Level)
				fmt.Printf("%s--- %s: %s (%s)\n", indent, summary.Status, summary.Name, summary.Time)
				if *listFailuresWithOutput {
					if lines, exists := testBuffers[summary.Name]; exists {
						for _, line := range lines {
							fmt.Printf("%s    %s\n", indent, line)
						}
					}
				}
			}
		}
		return
	}

	if *writeFiles {
		var dirsToCreate []string
		for testName := range testBuffers {
			dirsToCreate = append(dirsToCreate, filepath.Join(*outputDir, testName))
		}

		if !*forceFlag {
			for _, dirPath := range dirsToCreate {
				if _, err := os.Stat(dirPath); err == nil {
					fmt.Fprintf(os.Stderr, "Error: directory '%s' already exists.\n", dirPath)
					os.Exit(1)
				}
			}
		}

		for _, dirPath := range dirsToCreate {
			if err := os.MkdirAll(dirPath, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dirPath, err)
				os.Exit(1)
			}
			if *debugFlag {
				fmt.Printf("[DEBUG] Created directory: %s\n", dirPath)
			}
		}

		for testName, lines := range testBuffers {
			filePath := filepath.Join(*outputDir, testName, "output.log")
			file, err := os.Create(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file %s: %v\n", filePath, err)
				continue
			}
			if *debugFlag {
				fmt.Printf("[DEBUG] Writing to file: %s\n", filePath)
			}
			for _, line := range lines {
				_, _ = file.WriteString(line + "\n")
			}
			file.Close()
		}
		return
	}

	// Default behaviour: summarise failures only.
	for _, summary := range summaryRecords {
		if summary.Status == "FAIL" && reTest.MatchString(summary.Name) {
			indent := strings.Repeat("    ", summary.Level)
			fmt.Printf("%s--- %s: %s (%s)\n", indent, summary.Status, summary.Name, summary.Time)
		}
	}
}
