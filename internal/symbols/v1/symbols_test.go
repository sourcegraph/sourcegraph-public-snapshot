package v1

import (
	"encoding/json"
	"math/rand"
	reflect "reflect"
	"testing"
	"testing/quick"

	"github.com/sourcegraph/sourcegraph/internal/grpc/protorand"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func (sr *SearchRequest) Generate(rand *rand.Rand, size int) reflect.Value {
	return protorand.Generate[*SearchRequest](rand, size)
}

func TestXxx(t *testing.T) {
	err := quick.Check(func(sr *SearchRequest) bool {
		internal := sr.ToInternal()
		var sr2 SearchRequest
		sr2.FromInternal(&internal)
		if !proto.Equal(sr, &sr2) {
			m1, _ := json.Marshal(sr)
			m2, _ := json.Marshal(&sr2)
			require.Equal(t, m1, m2)
		}
		return true
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
}
