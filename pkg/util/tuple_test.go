package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Tuple1_Len(t *testing.T) {
	tup := Tuple1[int]{
		Val1: 1,
	}
	assert.Equal(t, 1, tup.Len())
}

func Test_Tuple1_Values(t *testing.T) {
	tup := Tuple1[int]{
		Val1: 1,
	}
	assert.Equal(t, []any{1}, tup.Values())
}

func Test_Tuple2_Len(t *testing.T) {
	tup := Tuple2[int, string]{
		Val1: 1,
		Val2: "two",
	}
	assert.Equal(t, 2, tup.Len())
}

func Test_Tuple2_Values(t *testing.T) {
	tup := Tuple2[int, string]{
		Val1: 1,
		Val2: "two",
	}
	assert.Equal(t, []any{1, "two"}, tup.Values())
}

func Test_Tuple3_Len(t *testing.T) {
	tup := Tuple3[int, string, bool]{
		Val1: 1,
		Val2: "two",
		Val3: true,
	}
	assert.Equal(t, 3, tup.Len())
}

func Test_Tuple3_Values(t *testing.T) {
	tup := Tuple3[int, string, bool]{
		Val1: 1,
		Val2: "two",
		Val3: true,
	}
	assert.Equal(t, []any{1, "two", true}, tup.Values())
}

func Test_Tuple4_Len(t *testing.T) {
	tup := Tuple4[int, string, bool, float64]{
		Val1: 1,
		Val2: "two",
		Val3: true,
		Val4: 4.0,
	}
	assert.Equal(t, 4, tup.Len())
}

func Test_Tuple4_Values(t *testing.T) {
	tup := Tuple4[int, string, bool, float64]{
		Val1: 1,
		Val2: "two",
		Val3: true,
		Val4: 4.0}
	assert.Equal(t, []any{1, "two", true, 4.0}, tup.Values())
}

func Test_Tuple5_Len(t *testing.T) {
	tup := Tuple5[int, string, bool, float64, rune]{
		Val1: 1,
		Val2: "two",
		Val3: true,
		Val4: 4.0,
		Val5: 'a',
	}
	assert.Equal(t, 5, tup.Len())
}

func Test_Tuple5_Values(t *testing.T) {
	tup := Tuple5[int, string, bool, float64, rune]{
		Val1: 1,
		Val2: "two",
		Val3: true,
		Val4: 4.0,
		Val5: 'a',
	}
	assert.Equal(t, []any{1, "two", true, 4.0, 'a'}, tup.Values())
}

func Test_Tuple6_Len(t *testing.T) {
	tup := Tuple6[int, string, bool, float64, rune, byte]{
		Val1: 1,
		Val2: "two",
		Val3: true,
		Val4: 4.0,
		Val5: 'a',
		Val6: byte(255),
	}
	assert.Equal(t, 6, tup.Len())
}

func Test_Tuple6_Values(t *testing.T) {
	tup := Tuple6[int, string, bool, float64, rune, byte]{
		Val1: 1,
		Val2: "two",
		Val3: true,
		Val4: 4.0,
		Val5: 'a',
		Val6: byte(255),
	}
	assert.Equal(t, []any{1, "two", true, 4.0, 'a', byte(255)}, tup.Values())
}
