package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

func TestRepoCloneProgress_ProtoRoundTrip(t *testing.T) {
	var diff string

	fn := func(original RepoCloneProgress) bool {
		var converted RepoCloneProgress
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("RepoCloneProgress proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestSearchRequestProtoRoundtrip(t *testing.T) {
	req := &SearchRequest{
		Repo:      "test1",
		Revisions: []string{"ABC"},
		Query: &Operator{
			Kind: And,
			Operands: []Node{
				&AuthorMatches{Expr: "abc", IgnoreCase: true},
				&CommitAfter{Time: time.Date(2021, 12, 3, 12, 3, 45, 0, time.UTC)},
			},
		},
		Limit: 42,
	}

	protoReq := req.ToProto()
	roundtripped, err := SearchRequestFromProto(protoReq)
	require.NoError(t, err)
	require.Equal(t, req, roundtripped)
}

func TestCommitMatchProtoRoundtrip(t *testing.T) {
	req := CommitMatch{
		Oid:        "8a8a8a88aa",
		Author:     Signature{Name: "sasha", Email: "cap@map.com", Date: time.Date(2022, 3, 4, 2, 3, 4, 0, time.UTC)},
		Committer:  Signature{Name: "mushu", Email: "lop@cop.com", Date: time.Date(2022, 3, 4, 2, 3, 4, 0, time.UTC)},
		Parents:    []api.CommitID{"9b9b9b", "2c22c2c"},
		Refs:       []string{"pale", "blue", "dot"},
		SourceRefs: []string{"giant", "red", "spot"},
		Message: result.MatchedString{
			Content: "lorem ipsum",
			MatchedRanges: result.Ranges{{
				Start: result.Location{Offset: 10, Line: 4, Column: 2},
				End:   result.Location{Offset: 11, Line: 5, Column: 3},
			}},
		},
		Diff: result.MatchedString{
			Content: "dolor",
			MatchedRanges: result.Ranges{{
				Start: result.Location{Offset: 888, Line: 999, Column: 444},
				End:   result.Location{Offset: 111, Line: 222, Column: 333},
			}},
		},
	}

	protoReq := req.ToProto()
	roundtripped := CommitMatchFromProto(protoReq)
	require.Equal(t, req, roundtripped)
}

func TestGetObjectRequestProtoRoundtrip(t *testing.T) {
	var diff string

	fn := func(original GetObjectRequest) bool {
		protoReq := original.ToProto()

		var converted GetObjectRequest
		converted.FromProto(protoReq)

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("GetObjectRequest proto roundtrip failed (-want +got):\n%s", diff)
	}
}

func TestGetObjectResponseProtoRoundtrip(t *testing.T) {
	var diff string

	fn := func(id [20]byte, typ fuzzObjectType) bool {
		original := GetObjectResponse{
			Object: gitdomain.GitObject{
				ID:   id,
				Type: gitdomain.ObjectType(typ),
			},
		}
		protoResp := original.ToProto()

		var converted GetObjectResponse
		converted.FromProto(protoResp)

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(fn, nil); err != nil {
		t.Errorf("GetObjectResponse proto roundtrip failed (-want +got):\n%s", diff)
	}
}

type fuzzObjectType gitdomain.ObjectType

func (fuzzObjectType) Generate(r *rand.Rand, _ int) reflect.Value {
	validValues := []gitdomain.ObjectType{gitdomain.ObjectTypeCommit, gitdomain.ObjectTypeTag, gitdomain.ObjectTypeTree, gitdomain.ObjectTypeBlob}
	return reflect.ValueOf(fuzzObjectType(validValues[r.Intn(len(validValues))]))
}

var _ quick.Generator = fuzzObjectType(gitdomain.ObjectTypeCommit)
