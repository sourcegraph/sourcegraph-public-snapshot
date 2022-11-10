package ranking

import (
	"context"
)

type streamingGraph interface {
	Next(ctx context.Context) (from, to string, ok bool, err error)
}

type graphStreamer struct {
	ch <-chan streamedEdge
}

type streamedEdge struct {
	from, to string
	err      error
}

func (gs *graphStreamer) Next(ctx context.Context) (from, to string, ok bool, err error) {
	select {
	case p, ok := <-gs.ch:
		if !ok {
			return "", "", false, nil
		}

		return p.from, p.to, true, p.err

	case <-ctx.Done():
		return "", "", false, ctx.Err()
	}
}
