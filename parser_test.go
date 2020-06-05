package tap13

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkParsingSpeed(b *testing.B) {
	lines := ReadFile("testdata/edge_cases.tap13")
	for i := 0; i < b.N; i++ {
		Parse(lines)
	}
}

func TestParseResults(t *testing.T) {
	t.Run("NoInputFails", func(t *testing.T) {
		input := []string{}
		result := Parse(input)
		assert.False(t, result.IsPassing())
	})
	t.Run("SingleFailingTestShouldFail", func(t *testing.T) {
		input := strings.Split(`TAP version 13
not ok`,
			"\n")
		result := Parse(input)
		assert.False(t, result.IsPassing())
	})
	t.Run("SinglePassingTestShouldPass", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok`,
			"\n")
		result := Parse(input)
		assert.True(t, result.IsPassing())
	})
	t.Run("PassFailMixtureShouldFail", func(t *testing.T) {
		input := strings.Split(`TAP version 13
ok
not ok
ok`,
			"\n")
		result := Parse(input)
		assert.False(t, result.IsPassing())
		assert.Equal(t, 3, result.TotalTests)
	})

}
