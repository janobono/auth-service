package util

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestIsBlank(t *testing.T) {
	type testCase struct {
		value    string
		expected bool
	}

	t.Run("IsBlank", func(t *testing.T) {
		tests := []testCase{
			{value: "", expected: true},
			{value: "  ", expected: true},
			{value: "anything", expected: false},
		}

		for _, test := range tests {
			actual := IsBlank(test.value)
			assert.Equal(t, actual, test.expected)
		}
	})
}
