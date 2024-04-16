package service

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Test_copyBlobs tests that we concatenate objects from the object store
// properly.
func Test_copyBlobs(t *testing.T) {
	keysIter := iterator.From([]string{"a", "b", "c"})

	blobs := map[string]io.Reader{
		"a": bytes.NewReader([]byte("{\"Key\":\"a\"}\n{\"Key\":\"b\"}\n")),
		"b": bytes.NewReader([]byte("{\"Key\":\"c\"}\n")),
		"c": bytes.NewReader([]byte("{\"Key\":\"d\"}\n{\"Key\":\"e\"}\n{\"Key\":\"f\"}\n")),
	}

	blobstore := mocks.NewMockStore()
	blobstore.GetFunc.SetDefaultHook(func(ctx context.Context, key string) (io.ReadCloser, error) {
		return io.NopCloser(blobs[key]), nil
	})

	w := &bytes.Buffer{}

	n, err := writeSearchJobJSON(context.Background(), keysIter, blobstore, w)
	require.NoError(t, err)
	require.Equal(t, int64(72), n)

	want := "{\"Key\":\"a\"}\n{\"Key\":\"b\"}\n{\"Key\":\"c\"}\n{\"Key\":\"d\"}\n{\"Key\":\"e\"}\n{\"Key\":\"f\"}\n"
	require.Equal(t, want, w.String())
}

func TestIsEnabled(t *testing.T) {
	defer conf.Mock(nil)

	enabled := true
	disabled := false

	cases := []struct {
		name                 string
		experimentalFeatures *schema.ExperimentalFeatures
		want                 bool
	}{
		{
			name:                 "explicitly enabled",
			experimentalFeatures: &schema.ExperimentalFeatures{SearchJobs: &enabled},
			want:                 true,
		},
		{
			name:                 "ExperimentalFeatures=nil",
			experimentalFeatures: nil,
			want:                 true,
		},
		{
			name:                 "SearchJobs=nil",
			experimentalFeatures: &schema.ExperimentalFeatures{},
			want:                 true,
		},
		{
			name:                 "explicitly disabled",
			experimentalFeatures: &schema.ExperimentalFeatures{SearchJobs: &disabled},
			want:                 false,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: c.experimentalFeatures}})
			require.Equal(t, c.want, isEnabled())
		})
	}
}
