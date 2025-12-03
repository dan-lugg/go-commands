package util

type Tuple interface {
	Len() int
	Values() []any
}

type Tuple1[T1 any] struct {
	Val1 T1
}

func (t Tuple1[T1]) Len() int {
	return 1
}

func (t Tuple1[T1]) Values() []any {
	return []any{t.Val1}
}

type Tuple2[T1 any, T2 any] struct {
	Val1 T1
	Val2 T2
}

func (t Tuple2[T1, T2]) Len() int {
	return 2
}

func (t Tuple2[T1, T2]) Values() []any {
	return []any{t.Val1, t.Val2}
}

type Tuple3[T1 any, T2 any, T3 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
}

func (t Tuple3[T1, T2, T3]) Len() int {
	return 3
}

func (t Tuple3[T1, T2, T3]) Values() []any {
	return []any{t.Val1, t.Val2, t.Val3}
}

type Tuple4[T1 any, T2 any, T3 any, T4 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
}

func (t Tuple4[T1, T2, T3, T4]) Len() int {
	return 4
}

func (t Tuple4[T1, T2, T3, T4]) Values() []any {
	return []any{t.Val1, t.Val2, t.Val3, t.Val4}
}

type Tuple5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
	Val5 T5
}

func (t Tuple5[T1, T2, T3, T4, T5]) Len() int {
	return 5
}

func (t Tuple5[T1, T2, T3, T4, T5]) Values() []any {
	return []any{t.Val1, t.Val2, t.Val3, t.Val4, t.Val5}
}

type Tuple6[T1 any, T2 any, T3 any, T4 any, T5 any, T6 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
	Val4 T4
	Val5 T5
	Val6 T6
}

func (t Tuple6[T1, T2, T3, T4, T5, T6]) Len() int {
	return 6
}

func (t Tuple6[T1, T2, T3, T4, T5, T6]) Values() []any {
	return []any{t.Val1, t.Val2, t.Val3, t.Val4, t.Val5, t.Val6}
}
