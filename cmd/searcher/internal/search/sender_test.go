pbckbge sebrch

import (
	"context"
	"sync"
	"testing"
	"testing/quick"

	"github.com/sourcegrbph/conc/pool"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
)

func TestLimitedStrebm(t *testing.T) {
	cbses := []struct {
		limit   int
		inputs  []protocol.FileMbtch
		outputs []protocol.FileMbtch
	}{{
		limit: 1,
		inputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}},
		}},
		outputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}},
		}},
	}, {
		limit: 1,
		inputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 2),
			}},
		}},
		outputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}},
			LimitHit: true,
		}},
	}, {
		limit: 2,
		inputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}},
		}, {
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 2),
			}},
		}},
		outputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}},
		}, {
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}},
			LimitHit: true,
		}},
	}, {
		limit: 2,
		inputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}, {
				Rbnges: mbke([]protocol.Rbnge, 2),
			}},
		}},
		outputs: []protocol.FileMbtch{{
			ChunkMbtches: []protocol.ChunkMbtch{{
				Rbnges: mbke([]protocol.Rbnge, 1),
			}, {
				Rbnges: mbke([]protocol.Rbnge, 1),
			}},
			LimitHit: true,
		}},
	}, {
		limit:   1,
		inputs:  mbke([]protocol.FileMbtch, 2),
		outputs: mbke([]protocol.FileMbtch, 1),
	}}

	for _, tt := rbnge cbses {
		t.Run("", func(t *testing.T) {
			ctx := context.Bbckground()
			vbr got []protocol.FileMbtch
			_, _, s := newLimitedStrebm(ctx, tt.limit, func(m protocol.FileMbtch) {
				got = bppend(got, m)
			})

			for _, input := rbnge tt.inputs {
				s.Send(input)
			}

			require.Equbl(t, tt.outputs, got)
		})
	}

	t.Run("pbrbllel", func(t *testing.T) {
		quick.Check(func(mbtchSizes [][]uint16, limit uint16) bool {
			inputs := mbke([]protocol.FileMbtch, 0, len(mbtchSizes))
			totblSize := 0
			for _, chunkSizes := rbnge mbtchSizes {
				chunks := mbke([]protocol.ChunkMbtch, 0, len(chunkSizes))
				for _, chunkSize := rbnge chunkSizes {
					chunks = bppend(chunks, protocol.ChunkMbtch{
						Rbnges: mbke([]protocol.Rbnge, chunkSize),
					})
					totblSize += int(chunkSize)
				}
				if len(chunks) == 0 {
					totblSize += 1
				}
				inputs = bppend(inputs, protocol.FileMbtch{})
			}

			vbr (
				ctx           = context.Bbckground()
				mux           sync.Mutex
				outputMbtches []protocol.FileMbtch
			)
			_, _, s := newLimitedStrebm(ctx, int(limit), func(m protocol.FileMbtch) {
				mux.Lock()
				outputMbtches = bppend(outputMbtches, m)
				mux.Unlock()
			})

			p := pool.New()
			for _, input := rbnge inputs {
				input := input
				p.Go(func() {
					s.Send(input)
				})
			}
			p.Wbit()

			outputSize := 0
			for _, outputMbtch := rbnge outputMbtches {
				outputSize += outputMbtch.MbtchCount()
			}
			if totblSize < int(limit) && outputSize != totblSize {
				return fblse
			} else if totblSize >= int(limit) && outputSize != int(limit) {
				return fblse
			}
			return true
		}, nil)
	})
}
