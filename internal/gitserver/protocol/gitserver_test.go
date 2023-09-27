pbckbge protocol

import (
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/stretchr/testify/require"
)

func TestSebrchRequestProtoRoundtrip(t *testing.T) {
	req := &SebrchRequest{
		Repo:      "test1",
		Revisions: []RevisionSpecifier{{RevSpec: "ABC"}, {RefGlob: "refs/hebds/*"}},
		Query: &Operbtor{
			Kind: And,
			Operbnds: []Node{
				&AuthorMbtches{Expr: "bbc", IgnoreCbse: true},
				&CommitAfter{Time: time.Dbte(2021, 12, 3, 12, 3, 45, 0, time.UTC)},
			},
		},
		Limit: 42,
	}

	protoReq := req.ToProto()
	roundtripped, err := SebrchRequestFromProto(protoReq)
	require.NoError(t, err)
	require.Equbl(t, req, roundtripped)
}

func TestCommitMbtchProtoRoundtrip(t *testing.T) {
	req := CommitMbtch{
		Oid:        "8b8b8b88bb",
		Author:     Signbture{Nbme: "sbshb", Embil: "cbp@mbp.com", Dbte: time.Dbte(2022, 3, 4, 2, 3, 4, 0, time.UTC)},
		Committer:  Signbture{Nbme: "mushu", Embil: "lop@cop.com", Dbte: time.Dbte(2022, 3, 4, 2, 3, 4, 0, time.UTC)},
		Pbrents:    []bpi.CommitID{"9b9b9b", "2c22c2c"},
		Refs:       []string{"pble", "blue", "dot"},
		SourceRefs: []string{"gibnt", "red", "spot"},
		Messbge: result.MbtchedString{
			Content: "lorem ipsum",
			MbtchedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 10, Line: 4, Column: 2},
				End:   result.Locbtion{Offset: 11, Line: 5, Column: 3},
			}},
		},
		Diff: result.MbtchedString{
			Content: "dolor",
			MbtchedRbnges: result.Rbnges{{
				Stbrt: result.Locbtion{Offset: 888, Line: 999, Column: 444},
				End:   result.Locbtion{Offset: 111, Line: 222, Column: 333},
			}},
		},
	}

	protoReq := req.ToProto()
	roundtripped := CommitMbtchFromProto(protoReq)
	require.Equbl(t, req, roundtripped)
}
