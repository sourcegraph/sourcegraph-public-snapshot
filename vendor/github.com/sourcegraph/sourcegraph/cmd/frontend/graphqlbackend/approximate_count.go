package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

func newApproximateCount(limit int32, f func(limit int32) (int32, error)) (*approximateCount, error) {
	n, err := f(limit + 1)
	timedOut := err == context.DeadlineExceeded || errors.Cause(err) == context.DeadlineExceeded
	if err != nil && !timedOut {
		return nil, err
	}
	c := &approximateCount{
		count: n,
		exact: err == nil,
	}
	if c.count > limit {
		c.count = limit
		c.exact = false
	}
	return c, nil
}

type approximateCount struct {
	count int32
	exact bool
}

func (r *approximateCount) Count() int32 { return r.count }
func (r *approximateCount) Exact() bool  { return r.exact }
func (r *approximateCount) Label() string {
	if r.exact {
		return strconv.Itoa(int(r.count))
	}
	return fmt.Sprintf("%d+", r.count)
}
