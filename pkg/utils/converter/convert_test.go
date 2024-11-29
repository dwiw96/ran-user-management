package converter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertStrToInt(t *testing.T) {
	tests := []struct {
		input string
		ans   int
	}{
		{
			input: "0",
			ans:   0,
		}, {
			input: "50",
			ans:   50,
		}, {
			input: "324",
			ans:   324,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			res, err := ConvertStrToInt(test.input)
			require.NoError(t, err)
			assert.Equal(t, test.ans, res)
		})
	}
}

func TestConvertStrToInt32(t *testing.T) {
	testCases := []struct {
		input string
		ans   int32
		isErr bool
	}{
		{
			input: "0",
			ans:   0,
			isErr: false,
		}, {
			input: "50",
			ans:   50,
			isErr: false,
		}, {
			input: "324",
			ans:   324,
			isErr: false,
		}, {
			input: "-324",
			ans:   -324,
			isErr: false,
		}, {
			input: "32a4",
			ans:   -1,
			isErr: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			res, err := ConvertStrToInt32(tC.input)
			if !tC.isErr {
				require.NoError(t, err)
				assert.Equal(t, tC.ans, res)
			} else {
				require.Error(t, err)
				assert.Equal(t, tC.ans, res)
			}
		})
	}
}
