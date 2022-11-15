package syncjobs

import (
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSyncJobsRecordsStoreWatch(t *testing.T) {
	s := NewRecordsStore(logtest.Scoped(t))

	// assert default
	assert.IsType(t, noopCache{}, s.cache)

	cw := confWatcher{
		conf: schema.SiteConfiguration{
			AuthzSyncJobsRecordsTTL: 5,
		},
	}

	// register
	s.Watch(&cw)

	// proc the update
	cw.update()

	// assert updated
	assert.Equal(t, 5*time.Minute, s.cache.(*rcache.Cache).TTL())
}

func TestSyncJobRecordsRecord(t *testing.T) {
	mockTime, err := time.Parse(time.RFC1123, time.RFC1123)
	if err != nil {
		t.Fatal(err.Error())
	}
	s := RecordsStore{
		logger: logtest.Scoped(t),
		now:    func() time.Time { return mockTime },
	}
	t.Run("success", func(t *testing.T) {
		c := memCache{}
		s.cache = c
		s.Record("repo", 12, []ProviderStatus{}, nil)
		autogold.Want("record_success", memCache{"1136214245000000000": `{"job_type":"repo","job_id":12,"completed":"2006-01-02T15:04:05Z","status":"SUCCESS","message":"","providers":[]}`}).
			Equal(t, c)
	})
	t.Run("error", func(t *testing.T) {
		c := memCache{}
		s.cache = c
		s.Record("repo", 12, []ProviderStatus{}, errors.New("oh no"))
		autogold.Want("record_error", memCache{"1136214245000000000": `{"job_type":"repo","job_id":12,"completed":"2006-01-02T15:04:05Z","status":"ERROR","message":"oh no","providers":[]}`}).
			Equal(t, c)
	})
}
