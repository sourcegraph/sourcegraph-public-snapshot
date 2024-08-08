package search

import (
	"context"
	"sync"
	"testing"
	"testing/quick"

	"github.com/sourcegraph/conc/pool"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/searcher/protocol"
)

func TestLimitedStream(t *testing.T) {
	cases := []struct {
		limit   int
		inputs  []protocol.FileMatch
		outputs []protocol.FileMatch
	}{{
		limit: 1,
		inputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}},
		}},
		outputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}},
		}},
	}, {
		limit: 1,
		inputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 2),
			}},
		}},
		outputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}},
			LimitHit: true,
		}},
	}, {
		limit: 2,
		inputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}},
		}, {
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 2),
			}},
		}},
		outputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}},
		}, {
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}},
			LimitHit: true,
		}},
	}, {
		limit: 2,
		inputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}, {
				Ranges: make([]protocol.Range, 2),
			}},
		}},
		outputs: []protocol.FileMatch{{
			ChunkMatches: []protocol.ChunkMatch{{
				Ranges: make([]protocol.Range, 1),
			}, {
				Ranges: make([]protocol.Range, 1),
			}},
			LimitHit: true,
		}},
	}, {
		limit:   1,
		inputs:  make([]protocol.FileMatch, 2),
		outputs: make([]protocol.FileMatch, 1),
	}}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			ctx := context.Background()
			var got []protocol.FileMatch
			_, _, s := newLimitedStream(ctx, tt.limit, func(m protocol.FileMatch) {
				got = append(got, m)
			})

			for _, input := range tt.inputs {
				s.Send(input)
			}

			require.Equal(t, tt.outputs, got)
		})
	}

	t.Run("parallel", func(t *testing.T) {
		quick.Check(func(matchSizes [][]uint16, limit uint16) bool {
			inputs := make([]protocol.FileMatch, 0, len(matchSizes))
			totalSize := 0
			for _, chunkSizes := range matchSizes {
				chunks := make([]protocol.ChunkMatch, 0, len(chunkSizes))
				for _, chunkSize := range chunkSizes {
					chunks = append(chunks, protocol.ChunkMatch{
						Ranges: make([]protocol.Range, chunkSize),
					})
					totalSize += int(chunkSize)
				}
				if len(chunks) == 0 {
					totalSize += 1
				}
				inputs = append(inputs, protocol.FileMatch{})
			}

			var (
				ctx           = context.Background()
				mux           sync.Mutex
				outputMatches []protocol.FileMatch
			)
			_, _, s := newLimitedStream(ctx, int(limit), func(m protocol.FileMatch) {
				mux.Lock()
				outputMatches = append(outputMatches, m)
				mux.Unlock()
			})

			p := pool.New()
			for _, input := range inputs {
				input := input
				p.Go(func() {
					s.Send(input)
				})
			}
			p.Wait()

			outputSize := 0
			for _, outputMatch := range outputMatches {
				outputSize += outputMatch.MatchCount()
			}
			if totalSize < int(limit) && outputSize != totalSize {
				return false
			} else if totalSize >= int(limit) && outputSize != int(limit) {
				return false
			}
			return true
		}, nil)
	})
}
