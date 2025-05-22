package util

import (
	"gotest.tools/v3/assert"
	"testing"
)

func TestCompare(t *testing.T) {
	type testCase struct {
		password        string
		encodedPassword string
		expected        bool
	}

	t.Run("ToScDf", func(t *testing.T) {
		simple, err := Encode("simple")
		if err != nil {
			t.Fatalf("Error creating password: %s", err)
		}

		tests := []testCase{
			{password: "simple", encodedPassword: "simple", expected: false},
			{password: "simple", encodedPassword: simple, expected: true},
		}

		for _, test := range tests {
			err := Compare(test.password, test.encodedPassword)
			if test.expected {
				assert.NilError(t, err)
			} else {
				assert.Error(t, err, "crypto/bcrypt: hashedSecret too short to be a bcrypted password")
			}
		}
	})
}
