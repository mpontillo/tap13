package tap13

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Test struct {
	TestNumber int
	Passed bool
	Failed bool
	Skipped bool
	Todo bool
	Description string
	Diagnostics []string
}

type Results struct {
	ExpectedTests int
	TotalTests int
	PassedTests int
	FailedTests int
	SkippedTests int
	TodoTests int
	TapVersion int
	Tests []Test
	Lines []string
	Explanation []string
}

const (
	FIND_VERSION_STRING = iota
	STORE_TEST_METADATA
	STORE_YAML
)

func (r* Results) String() string {
	var result string = ""
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
		result += fmt.Sprintf("  Missing tests: %d\n", r.ExpectedTests - r.TotalTests)
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
	return result
}

func (r *Results) IsPassing() bool {
	var testCount int
	if r.ExpectedTests > 0 {
		// For a planned run, we must ensure that the number of passing tests is equal to the
		// number of tests in the plan. Otherwise, at least one test was missing, and the run
		// should be considered a failure.
		testCount = r.ExpectedTests
	} else {
		// If the test run was "unplanned", the total number of tests is not known, so we must
		// assume that the total number of tests is equal to the number of tests that were found.
		testCount = r.TotalTests
	}
	return r.TodoTests + r.SkippedTests + r.PassedTests == testCount
}

var versionLine = regexp.MustCompile(`^TAP version (\d+)`)
var testLine = regexp.MustCompile(`^(not )?ok\b(.*)`)
var optionalTestLine = regexp.MustCompile(`\s*(\d*)?\s*([^#]*)(#\s*(\w*)\s*.*)?`)
var testPlanDeclaration = regexp.MustCompile(`^\d+\.\.(\d+)$`)
var diagnostic = regexp.MustCompile(`\s*#(.*)$`)

func Parse(lines []string) *Results {
	var err error
	var currentTest *Test
	state := FIND_VERSION_STRING
	foundTestPlan := false
	foundAllTests := false
	results := &Results{
		Lines: lines,
	}
	for _, line := range lines {
		switch state {
		case FIND_VERSION_STRING:
			versionMatch := versionLine.FindStringSubmatch(line)
			if versionMatch != nil {
				results.TapVersion, err = strconv.Atoi(versionMatch[1])
				if err != nil {
					// malformed test version line; keep looking
					continue
				}
				state = STORE_TEST_METADATA
			}
		case STORE_TEST_METADATA:
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
				directive := optionalContentMatch[4]
				testNumString := optionalContentMatch[1]
				if testNumString != "" {
					currentTest.TestNumber, err = strconv.Atoi(testNumString)
					if err != nil {
						currentTest.TestNumber = 0
					}
				}
				description := optionalContentMatch[2]
				currentTest.Description = description
				isFailed := testLineMatch[1] == "not "
				// Process special cases first; they should not count toward the pass/fail count.
				results.TotalTests++
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
			} else if strings.TrimSpace(line) == "---" {
				state = STORE_YAML
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
		case STORE_YAML:
			if strings.TrimSpace(line) == "..." {
				state = STORE_TEST_METADATA
				continue
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

func ReadFile(name string) []string {
	file, err := os.Open(name)

	if err != nil {
		log.Fatalf("Could not open file: %s", err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}
