package ranking

import (
	"context"

	"github.com/dcadenas/pagerank"
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

const pageRankFollowProbability = 0.85 // random jump 15% of the time
const pageRankTolerance = 0.0001

func (s *Service) pageRankFromStreamingGraph(ctx context.Context, graph streamingGraph) (map[string][]float64, error) {
	g := pagerank.New()
	idsToName := map[int]string{}
	nameToIDs := map[string]int{}

	for {
		from, to, ok, err := graph.Next(ctx)
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}

		fromID, ok := nameToIDs[from]
		if !ok {
			fromID = len(idsToName)
			idsToName[fromID] = from
			nameToIDs[from] = fromID
		}

		toID, ok := nameToIDs[to]
		if !ok {
			toID = len(idsToName)
			idsToName[toID] = from
			nameToIDs[from] = toID
		}

		g.Link(fromID, toID)
	}

	ranks := map[string][]float64{}
	g.Rank(pageRankFollowProbability, pageRankTolerance, func(identifier int, rank float64) {
		ranks[idsToName[identifier]] = []float64{1 - rank}
	})

	return ranks, nil
}
