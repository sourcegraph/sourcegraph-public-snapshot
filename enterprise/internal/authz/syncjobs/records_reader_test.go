package syncjobs

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSyncJobRecordsRead(t *testing.T) {
	c := memCache{}

	// Write multiple records
	s := NewRecordsStore(logtest.Scoped(t))
	s.cache = c
	s.Record("repo", 12, []ProviderStatus{{
		ProviderID:   "https://github.com",
		ProviderType: "github",
	}}, errors.New("oh no"))
	s.Record("repo", 15, []ProviderStatus{{
		ProviderID:   "https://github.com",
		ProviderType: "github",
	}}, nil)
	s.Record("user", 6, []ProviderStatus{{
		ProviderID:   "https://github.com",
		ProviderType: "github",
	}}, nil)

	// set up reader
	r := NewRecordsReader()
	r.readOnlyCache = c

	t.Run("read limited", func(t *testing.T) {
		results, err := r.GetAll(context.Background(), 1)
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		first := results[0]
		assert.Equal(t, "repo", first.JobType)
		assert.Equal(t, int32(12), first.JobID)
		assert.Len(t, first.Providers, 1)
	})

	t.Run("read all", func(t *testing.T) {
		results, err := r.GetAll(context.Background(), 10)
		assert.NoError(t, err)
		assert.Len(t, results, 3)

		// Assert sorted
		first := results[0]
		second := results[1]
		third := results[2]
		assert.True(t, first.Completed.Before(second.Completed))
		assert.True(t, second.Completed.Before(third.Completed))

		t.Run("read single", func(t *testing.T) {
			s, err := r.Get(second.Completed)
			assert.NoError(t, err)
			assert.Equal(t, second, *s)
		})
	})

}
