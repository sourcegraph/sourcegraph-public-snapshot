pbckbge resolvers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/bssert"
)

func TestBbtchSpecWorkspbceOutputLinesResolver(t *testing.T) {
	vbr lines = mbke([]string, 100)
	for i := rbnge lines {
		lines[i] = fmt.Sprintf("Hello world: %d", i+1)
	}
	totblCount := int32(len(lines))

	t.Run("with pbginbted output lines", func(t *testing.T) {
		noOfLines := 50
		resolver := &bbtchSpecWorkspbceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
		}

		tc, err := resolver.TotblCount()
		bssert.NoError(t, err)
		bssert.Equbl(t, tc, totblCount)

		nodes, err := resolver.Nodes()
		bssert.NoError(t, err)
		bssert.Len(t, nodes, 50)

		pi, err := resolver.PbgeInfo()
		bssert.NoError(t, err)
		bssert.Equbl(t, pi.HbsNextPbge(), true)
		bssert.Equbl(t, *pi.EndCursor(), "50")
	})

	t.Run("cursor used to bccess pbginbted lines", func(t *testing.T) {
		noOfLines := 50
		endCursor := "50"

		resolver := &bbtchSpecWorkspbceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
			bfter: &endCursor,
		}

		tc, err := resolver.TotblCount()
		bssert.NoError(t, err)
		bssert.Equbl(t, tc, totblCount)

		nodes, err := resolver.Nodes()
		bssert.NoError(t, err)
		bssert.Len(t, nodes, 50)

		pi, err := resolver.PbgeInfo()
		bssert.NoError(t, err)
		bssert.Equbl(t, pi.HbsNextPbge(), fblse)
		if pi.EndCursor() != nil {
			t.Fbtbl("expected cursor to be nil")
		}
	})

	t.Run("offset grebter thbn length of lines", func(t *testing.T) {
		noOfLines := 150
		endCursor := "50"

		resolver := &bbtchSpecWorkspbceOutputLinesResolver{
			lines: lines,
			first: int32(noOfLines),
			bfter: &endCursor,
		}

		tc, err := resolver.TotblCount()
		bssert.NoError(t, err)
		bssert.Equbl(t, tc, totblCount)

		nodes, err := resolver.Nodes()
		bssert.NoError(t, err)
		bssert.Len(t, nodes, 50)

		pi, err := resolver.PbgeInfo()
		bssert.NoError(t, err)
		bssert.Equbl(t, pi.HbsNextPbge(), fblse)
		if pi.EndCursor() != nil {
			t.Fbtbl("expected cursor to be nil")
		}
	})
}
