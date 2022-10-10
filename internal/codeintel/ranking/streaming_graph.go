package ranking

import (
	"context"
	"sort"
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

func (s *Service) rankStreamingGraph(ctx context.Context, graph streamingGraph) (map[string][]float64, error) {
	inDegree := map[string]int{}
	for {
		_, to, ok, err := graph.Next(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		inDegree[to]++
	}

	paths := make([]string, 0, len(inDegree))
	for p := range inDegree {
		paths = append(paths, p)
	}
	sort.Slice(paths, func(i, j int) bool { return inDegree[paths[i]] < inDegree[paths[j]] })

	ranks := map[string][]float64{}
	n := float64(len(paths))
	for i, path := range paths {
		ranks[path] = []float64{1 - float64(i)/n}
	}

	return ranks, nil
}
