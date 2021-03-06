package tap13

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	util "github.com/mpontillo/tap13/internal"
)

func BenchmarkParsingSpeed(b *testing.B) {
	lines := util.ReadFile("testdata/edge_cases.tap13")
	for i := 0; i < b.N; i++ {
		Parse(lines)
	}
}

func TestParseResults(t *testing.T) {
	t.Run("NoInputFails", func(t *testing.T) {
		var input []string
		result := Parse(input)
		assert.False(t, result.IsPassing())
		assert.Equal(t, ` Overall result: FAIL
Total tests run: 0
`, result.String())
	})
	t.Run("SingleFailingTestShouldFail", func(t *testing.T) {
		input := strings.Split(`TAP version 13
not ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.False(t, result.IsPassing())
		assert.Equal(t, ` Overall result: FAIL
Total tests run: 1
   Failed tests: 1
`, result.String())
	})
	t.Run("SinglePassingTestShouldPass", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 1
`, result.String())
	})
	t.Run("PassFailMixtureShouldFail", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok
not ok
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.False(t, result.IsPassing())
		assert.Equal(t, 3, result.TotalTests)
	})
	t.Run("MissingTestsShouldFail", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..4
ok
ok
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.False(t, result.IsPassing())
		assert.Equal(t, 3, result.TotalTests)
		assert.Equal(t, 4, result.ExpectedTests)
		assert.Equal(t, ` Overall result: FAIL
 Expected tests: 4
  Missing tests: 1
   Passed tests: 3
`, result.String())
	})
	t.Run("ExtraTestsShouldBeIgnored", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..4
ok
ok
ok
ok
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 4, result.TotalTests)
		assert.Equal(t, 4, result.ExpectedTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 4
`, result.String())
	})
	t.Run("SkipAndTodoTestsShouldBeIgnoredButCountedAndDirectivesSaved", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..5
ok
not ok # skip because we feel like skipping
ok # SKIP
not ok # todo
ok # TODO working on this one`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 5, result.TotalTests)
		assert.Equal(t, 5, result.ExpectedTests)
		assert.Equal(t, "", result.Tests[0].DirectiveText)
		assert.Equal(t,
			"skip because we feel like skipping",
			result.Tests[1].DirectiveText)
		assert.Equal(t,
			"TODO working on this one",
			result.Tests[4].DirectiveText)
		assert.Equal(t, ` Overall result: PASS
Total tests run: 5
   Passed tests: 1
  Skipped tests: 2
     TODO tests: 2
`, result.String())
	})
	t.Run("MalformedPlansShouldBeIgnored", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..N
1..400000000000000000000000000000000000000000000000000000000000000000000000000
1..4
1..M
ok
ok
ok
ok
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 4, result.TotalTests)
		assert.Equal(t, 4, result.ExpectedTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 4
`, result.String())
	})
	t.Run("TestsCanBeNumbered", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..4
ok 4
ok 2
ok 1
ok 3`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 4, result.TotalTests)
		assert.Equal(t, 4, result.ExpectedTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 4
`, result.String())
		assert.Equal(t, 4, result.Tests[0].TestNumber)
		assert.Equal(t, 2, result.Tests[1].TestNumber)
		assert.Equal(t, 1, result.Tests[2].TestNumber)
		assert.Equal(t, 3, result.Tests[3].TestNumber)
	})
	t.Run("CrazyTestNumbersRevertToNegativeOne", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..4
ok 400000000000000000000000000000000000000000000000000000000000000000000000000
ok 2
ok 1
ok 3`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 4, result.TotalTests)
		assert.Equal(t, 4, result.ExpectedTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 4
`, result.String())
		assert.Equal(t, -1, result.Tests[0].TestNumber)
		assert.Equal(t, 2, result.Tests[1].TestNumber)
		assert.Equal(t, 1, result.Tests[2].TestNumber)
		assert.Equal(t, 3, result.Tests[3].TestNumber)
	})

	t.Run("TestsCanHaveDescriptions", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..4
ok 1 foo
ok bar # todo
not ok baz # skip
ok foo bar # squirrel!`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, "foo", result.Tests[0].Description)
		assert.Equal(t, "bar", result.Tests[1].Description)
		assert.Equal(t, "baz", result.Tests[2].Description)
		assert.Equal(t, "foo bar", result.Tests[3].Description)
	})
	t.Run("SkipContentBeforeVersionAndSkipWackyVersions", func(t *testing.T) {
		input := strings.Split(`Test results:
TAP version 999999999999999999999999999999999999999999999999999999999999999999
TAP version 12
1..4
ok
ok
ok
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 4, result.TotalTests)
		assert.Equal(t, 4, result.ExpectedTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 4
`, result.String())
		assert.Equal(t, 12, result.TapVersion)
	})
	t.Run("DoesNotParseTestLinesWithinYamlBlocks", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok
ok
  ---
  extra_info: lots of it
not ok
ok
  ...
ok
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 4, result.TotalTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 4
`, result.String())
	})
	t.Run("StoresExplanationAndDiagnostics", func(t *testing.T) {
		input := strings.Split(`TAP version 13
1..4
#
# Tests infinite improbability
#
ok
ok
ok
#
# !! Warning: uncertain improbability.
#
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 4, result.TotalTests)
		assert.Equal(t, 4, result.ExpectedTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 4
`, result.String())
		assert.Equal(t, "Tests infinite improbability", result.Explanation[0])
		assert.Equal(
			t, "!! Warning: uncertain improbability.", result.Tests[2].Diagnostics[0])
	})
	t.Run("BailOutImmediately", func(t *testing.T) {
		input := strings.Split(`TAP version 13
Bail out!`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.False(t, result.IsPassing())
		assert.Equal(t, 0, result.TotalTests)
	})
	t.Run("BailOutWithWhitespaceAfterSomeTests", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok
ok
ok
Bail out! `,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.False(t, result.IsPassing())
		assert.Equal(t, 3, result.TotalTests)
		assert.Equal(t, "", result.BailOutReason)
		assert.Equal(t, ` Overall result: FAIL
   Passed tests: 3
     Bailed out: (no reason given)
`, result.String())
	})
	t.Run("BailOutWithReasonAfterSomeTests", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok
ok
ok
Bail out! Towel not found.`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.False(t, result.IsPassing())
		assert.Equal(t, 3, result.TotalTests)
		assert.Equal(t, "Towel not found.", result.BailOutReason)
		assert.Equal(t, ` Overall result: FAIL
   Passed tests: 3
     Bailed out: Towel not found.
`, result.String())
	})
	t.Run("StoresYamlBytes", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok
  ---
  yaml:
    foo: 1
    bar: 2
  ...
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 2, result.TotalTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 2
`, result.String())
		assert.Equal(t, []byte("  yaml:\n    foo: 1\n    bar: 2\n"), result.Tests[0].YamlBytes)
	})
	t.Run("StoresWeirdlyIndentedYamlBytes", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok
---
     yaml:
       foo: 1

       bar: 2
        ...
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.FoundTapData)
		assert.True(t, result.IsPassing())
		assert.Equal(t, 2, result.TotalTests)
		assert.Equal(t, ` Overall result: PASS
   Passed tests: 2
`, result.String())
		assert.Equal(t, []byte("     yaml:\n       foo: 1\n\n       bar: 2\n"),
			result.Tests[0].YamlBytes)
	})
	t.Run("InvalidInputFile", func(t *testing.T) {
		input := strings.Split(`Not a TAP version 13 file!
No TAP here.
What were you expecting? An error?`,
			"\n")
		result := Parse(input)
		assert.False(t, result.FoundTapData)
	})

}
