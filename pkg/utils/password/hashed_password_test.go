package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashingPassword(t *testing.T) {
	tests := []struct {
		name  string
		input string
		ans   string
		err   bool
	}{
		{
			name:  "success1",
			input: "password123",
			err:   false,
		}, {
			name:  "empty",
			input: "",
			err:   true,
		},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			res, err := HashingPassword(tests[i].input)
			if tests[i].err == false {
				require.NoError(t, err)
				assert.NotEmpty(t, res)

			} else {
				require.Error(t, err)
				assert.Empty(t, res)
				assert.Equal(t, "password is empty", err.Error())
			}
		})
	}
}

func TestVerifyHashPassword(t *testing.T) {
	tests := []struct {
		name   string
		input1 string
		input2 string
		err    bool
	}{
		{
			name:   "success",
			input1: "passMarkeT583",
			input2: "passMarkeT583",
			err:    false,
		}, {
			name:   "wrong",
			input1: "agVbnD97>?",
			input2: "agVbnD9>?",
			err:    true,
		},
	}

	for i := range tests {
		t.Run(tests[i].name, func(t *testing.T) {
			hashedPassword, err := HashingPassword(tests[i].input1)
			require.NoError(t, err)
			assert.NotEmpty(t, hashedPassword)
			err2 := VerifyHashPassword(tests[i].input2, hashedPassword)
			if tests[i].err == false {
				require.NoError(t, err2)
			} else {
				require.Error(t, err2)
			}
		})
	}
}
