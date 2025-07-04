package util

import "fmt"

var (
	ErrNotCataloged = fmt.Errorf("not cataloged")
)

type Option[TAny any] func(TAny)
