package syncjobs

import (
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
		c := &memCache{}
		s.cache = c
		s.Record("repo", 12, database.CodeHostStatusesSet{}, nil)
		autogold.Expect(&memCache{values: []string{
			`{"job_type":"repo","job_id":12,"completed":"2006-01-02T15:04:05Z","status":"SUCCESS","message":"","providers":[]}`,
		}}).
			Equal(t, c)
	})
	t.Run("error", func(t *testing.T) {
		c := &memCache{}
		s.cache = c
		s.Record("repo", 12, database.CodeHostStatusesSet{}, errors.New("oh no"))
		autogold.Expect(&memCache{values: []string{
			`{"job_type":"repo","job_id":12,"completed":"2006-01-02T15:04:05Z","status":"ERROR","message":"oh no","providers":[]}`,
		}}).
			Equal(t, c)
	})
}
