package git

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestDiffFileIterator(t *testing.T) {
	t.Run("Close", func(t *testing.T) {
		c := new(closer)
		i := &DiffFileIterator{rdr: c}
		i.Close()
		if *c != true {
			t.Errorf("iterator did not close the underlying reader: have: %v; want: %v", *c, true)
		}
	})
}

type closer bool

func (c *closer) Read(p []byte) (int, error) {
	return 0, errors.New("testing only; this should never be invoked")
}

func (c *closer) Close() error {
	*c = true
	return nil
}
