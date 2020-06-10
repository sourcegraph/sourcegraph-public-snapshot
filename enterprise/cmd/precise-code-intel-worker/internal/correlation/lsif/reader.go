package lsif

import (
	"context"
	"io"
)

type Pair struct {
	Element Element
	Err     error
}

type Reader func(ctx context.Context, r io.Reader) <-chan Pair
