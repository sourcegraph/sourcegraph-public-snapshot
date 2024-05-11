package protocol

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
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
