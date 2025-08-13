package util

type Option[TAny any] func(TAny)

type Tuple2[T1 any, T2 any] struct {
	Val1 T1
	Val2 T2
}

func NewTuple2[T1 any, T2 any](val1 T1, val2 T2) Tuple2[T1, T2] {
	return Tuple2[T1, T2]{
		Val1: val1,
		Val2: val2,
	}
}
