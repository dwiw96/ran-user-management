package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandomInt(t *testing.T) {
	tests := []struct {
		min int
		max int
	}{
		{
			min: 1,
			max: 5,
		}, {
			min: 5,
			max: 10,
		}, {
			min: 5,
			max: 15,
		}, {
			min: 70,
			max: 120,
		}, {
			min: 0,
			max: 100,
		}, {
			min: -500,
			max: -1,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			res := RandomInt(test.min, test.max)
			assert.GreaterOrEqual(t, res, test.min)
			assert.LessOrEqual(t, res, test.max)
		})
	}
}

func TestRandomInt32(t *testing.T) {
	testCases := []struct {
		desc string
		min  int32
		max  int32
	}{
		{
			desc: "success_positive",
			min:  1,
			max:  5,
		}, {
			desc: "success_positive",
			min:  5,
			max:  10,
		}, {
			desc: "success_positive",
			min:  5,
			max:  15,
		}, {
			desc: "success_positive",
			min:  70,
			max:  120,
		}, {
			desc: "success_positive",
			min:  0,
			max:  100,
		}, {
			desc: "success_negative",
			min:  -500,
			max:  -1,
		}, {
			desc: "success_negative",
			min:  -1000,
			max:  -500,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			res := RandomInt32(tC.min, tC.max)
			assert.GreaterOrEqual(t, res, tC.min)
			assert.LessOrEqual(t, res, tC.max)
		})
	}
}

func TestCreateRandomName(t *testing.T) {
	length := []int{3, 5, 7, 4, 2}
	for _, v := range length {
		res := CreateRandomString(v)
		require.NotEmpty(t, res)
		assert.Equal(t, v, len(res))
	}
}

func TestCreateRandomEmail(t *testing.T) {
	names := []string{"john", "doe", "test"}

	for _, v := range names {
		res := CreateRandomEmail(v)
		assert.Equal(t, res, v+"@mail.com")
	}
}
