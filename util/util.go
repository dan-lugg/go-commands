package util

type Option[TAny any] func(TAny)

type Tuple1[T1 any] struct {
	Val1 T1
}

type Tuple2[T1 any, T2 any] struct {
	Val1 T1
	Val2 T2
}

type Tuple3[T1 any, T2 any, T3 any] struct {
	Val1 T1
	Val2 T2
	Val3 T3
}
