package observation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
)

func TestMergeAttributes(t *testing.T) {
	type testCase struct {
		kv1  []attribute.KeyValue
		kv2  []attribute.KeyValue
		want []attribute.KeyValue
	}
	testCases := []testCase{
		{
			kv1: []attribute.KeyValue{
				attribute.Int("a", 0),
				attribute.Int("b", 1),
			},
			kv2: []attribute.KeyValue{
				attribute.Int("a", 1),
				attribute.Int("c", 2),
			},
			want: []attribute.KeyValue{
				attribute.Int("a", 1),
				attribute.Int("b", 1),
				attribute.Int("c", 2),
			},
		},
	}

	for _, tc := range testCases {
		require.Equal(t, tc.want, MergeAttributes(tc.kv1, tc.kv2...))
	}
}
