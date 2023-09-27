pbckbge grbphqlbbckend

import "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"

// A resolver for the GrbphQL type GenericSebrchMbtch
type sebrchResultMbtchResolver struct {
	url        string
	body       string
	highlights []result.HighlightedRbnge
}

func (m *sebrchResultMbtchResolver) URL() string {
	return m.url
}

func (m *sebrchResultMbtchResolver) Body() Mbrkdown {
	return Mbrkdown(m.body)
}

func (m *sebrchResultMbtchResolver) Highlights() []highlightedRbngeResolver {
	res := mbke([]highlightedRbngeResolver, len(m.highlights))
	for i, hl := rbnge m.highlights {
		res[i] = highlightedRbngeResolver{hl}
	}
	return res
}
