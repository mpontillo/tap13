/*
Package tap13 implements a parser for the Test Anything Protocol (TAP) version 13 specification.

The full protocol specification can be found at the following URL:

https://testanything.org/tap-version-13-specification.html

*/
package tap13

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Test encapsulates the result of a specific test, including a description and diagnostics (if
// supplied). The TestNumber field is undefined if the TAP output does not include test numbers.
// Diagnostics are supplied with trimmed whitespace, and blank lines removed.
type Test struct {
	TestNumber    int
	Passed        bool
	Failed        bool
	Skipped       bool
	Todo          bool
	Description   string
	DirectiveText string
	Diagnostics   []string
	YamlBytes     []byte
}

// Results encapsulates the result of the entire test run. If a plan was given in the input TAP, the
// ExpectedTests will be greater than or equal to zero. The input lines are preserved in the Lines
// field. Any diagnostics given before the output of a test run is preserved in the Explanation.
// The Tests field contains a Test struct for each test that was run, in the order that it appeared
// in the TAP output.
type Results struct {
	ExpectedTests int
	TotalTests    int
	PassedTests   int
	FailedTests   int
	SkippedTests  int
	TodoTests     int
	TapVersion    int
	BailOut       bool
	BailOutReason string
	Tests         []Test
	Lines         []string
	Explanation   []string
}

const (
	findVersionString = iota
	storeTestMetadata
	storeYaml
)

func (r *Results) String() string {
	var result = ""
	if r.IsPassing() {
		result += " Overall result: PASS\n"
	} else {
		result += " Overall result: FAIL\n"
	}
	if r.TotalTests == 0 || r.PassedTests != r.TotalTests {
		result += fmt.Sprintf("Total tests run: %d\n", r.TotalTests)
	}
	if r.ExpectedTests > 0 && r.ExpectedTests != r.TotalTests {
		result += fmt.Sprintf(" Expected tests: %d\n", r.ExpectedTests)
	}
	if r.ExpectedTests > 0 && r.TotalTests < r.ExpectedTests {
		result += fmt.Sprintf("  Missing tests: %d\n", r.ExpectedTests-r.TotalTests)
	}
	if r.PassedTests > 0 {
		result += fmt.Sprintf("   Passed tests: %d\n", r.PassedTests)
	}
	if r.FailedTests > 0 {
		result += fmt.Sprintf("   Failed tests: %d\n", r.FailedTests)
	}
	if r.SkippedTests > 0 {
		result += fmt.Sprintf("  Skipped tests: %d\n", r.SkippedTests)
	}
	if r.TodoTests > 0 {
		result += fmt.Sprintf("     TODO tests: %d\n", r.TodoTests)
	}
	if r.BailOut {
		var reason string
		if r.BailOutReason != "" {
			reason = r.BailOutReason
		} else {
			reason = "(no reason given)"
		}
		result += fmt.Sprintf("     Bailed out: %s\n", reason)
	}
	return result
}

// IsPassing checks if the test results should be considered passing (true) or failing (false).
func (r *Results) IsPassing() bool {
	if r.TapVersion < 0 {
		// We didn't find a TAP header, so we can't really call this a success.
		return false
	}
	if r.BailOut {
		return false
	}
	var testCount int
	if r.ExpectedTests >= 0 {
		// For a planned run, we must ensure that the number of passing tests is equal to the
		// number of tests in the plan. Otherwise, at least one test was missing, and the run
		// should be considered a failure.
		testCount = r.ExpectedTests
	} else {
		// If the test run was "unplanned", the total number of tests is not known, so we must
		// assume that the total number of tests is equal to the number of tests that were found.
		testCount = r.TotalTests
	}
	return r.TodoTests+r.SkippedTests+r.PassedTests == testCount
}

var versionLine = regexp.MustCompile(`^TAP version (\d+)`)
var bailOutLine = regexp.MustCompile(`^Bail out!\s*(\S.*)?$`)
var testLine = regexp.MustCompile(`^(not )?ok\b(.*)`)
var optionalTestLine = regexp.MustCompile(`\s*(\d*)?\s*([^#]*)(#\s*((\w*)\s*.*)\s*)?`)
var testPlanDeclaration = regexp.MustCompile(`^\d+\.\.(\d+)$`)
var diagnostic = regexp.MustCompile(`\s*#(.*)$`)
var yamlStart = regexp.MustCompile(`^\s*---$`)

// Parse interprets the specified lines as output lines from a program that generate TAP output,
// and returns a corresponding Results structure containing the test results based on its
// interpretation.
func Parse(lines []string) *Results {
	var err error
	var currentTest *Test
	var yamlStop = regexp.MustCompile(`^\s*\.\.\.$`)
	state := findVersionString
	foundTestPlan := false
	foundAllTests := false
	results := &Results{
		ExpectedTests: -1,
		TapVersion:    -1,
		Lines:         lines,
	}
	for _, line := range lines {
		switch state {
		case findVersionString:
			versionMatch := versionLine.FindStringSubmatch(line)
			if versionMatch != nil {
				results.TapVersion, err = strconv.Atoi(versionMatch[1])
				if err != nil {
					// malformed test version line; keep looking
					continue
				}
				state = storeTestMetadata
			}
		case storeTestMetadata:
			bailOutMatch := bailOutLine.FindStringSubmatch(line)
			if bailOutMatch != nil {
				results.BailOut = true
				results.BailOutReason = bailOutMatch[1]
				break
			}
			if !foundTestPlan {
				testPlan := testPlanDeclaration.FindStringSubmatch(line)
				if testPlan != nil {
					results.ExpectedTests, err = strconv.Atoi(testPlan[1])
					if err != nil {
						// malformed test plan; keep looking
						continue
					}
				}
			}
			testLineMatch := testLine.FindStringSubmatch(line)
			if testLineMatch != nil {
				// Store the one we were previously working with, and start a new one.
				if currentTest != nil {
					results.Tests = append(results.Tests, *currentTest)
				}
				currentTest = &Test{}
				if foundAllTests {
					// We've already found all the tests in the plan, so don't waste effort looking
					// for more. The only reason not to break here instead is because we might want
					// to parse any diagnostics following the test result output.
					continue
				}
				optionalContentMatch := optionalTestLine.FindStringSubmatch(testLineMatch[2])
				directive := optionalContentMatch[5]
				directiveText := optionalContentMatch[4]
				testNumString := optionalContentMatch[1]
				if testNumString != "" {
					currentTest.TestNumber, err = strconv.Atoi(testNumString)
					if err != nil {
						currentTest.TestNumber = -1
					}
				}
				description := strings.TrimSpace(optionalContentMatch[2])
				currentTest.Description = description
				isFailed := testLineMatch[1] == "not "
				// Process special cases first; they should not count toward the pass/fail count.
				results.TotalTests++
				if directive != "" {
					currentTest.DirectiveText = directiveText
				}
				if strings.EqualFold(directive, "skip") {
					results.SkippedTests++
					currentTest.Skipped = true
				} else if strings.EqualFold(directive, "todo") {
					results.TodoTests++
					currentTest.Todo = true
				} else if isFailed {
					results.FailedTests++
					currentTest.Failed = true
				} else {
					results.PassedTests++
					currentTest.Passed = true
				}
				if results.TotalTests == results.ExpectedTests {
					foundAllTests = true
				}
			} else if yamlStart.MatchString(line) {
				state = storeYaml
				continue
			} else {
				diagnosticMatch := diagnostic.FindStringSubmatch(line)
				if diagnosticMatch != nil {
					diagnosticLine := strings.TrimSpace(diagnosticMatch[1])
					if diagnosticLine == "" {
						continue
					}
					if currentTest != nil {
						currentTest.Diagnostics = append(currentTest.Diagnostics, diagnosticLine)
					} else {
						results.Explanation = append(results.Explanation, diagnosticLine)
					}
				}
			}
		case storeYaml:
			if yamlStop.MatchString(line) {
				state = storeTestMetadata
				continue
			} else {
				// YAML that appears before a test definition is undefined behavior.
				if currentTest != nil {
					// The Go YAML library expects a []byte, so store it that way for later usage.
					currentTest.YamlBytes = append(currentTest.YamlBytes, line...)
					currentTest.YamlBytes = append(currentTest.YamlBytes, "\n"...)
				}
			}
		}
	}
	// if we have a currentTest at this point, it hasn't been saved to the results yet,
	// since we weren't sure if an upcoming line would have been relevant to it or not.
	if currentTest != nil {
		results.Tests = append(results.Tests, *currentTest)
	}
	return results
}
