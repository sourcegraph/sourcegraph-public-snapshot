package authz

import (
	"testing"
	"time"

	"github.com/hexops/autogold"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type confWatcher struct {
	update func()
	conf   schema.SiteConfiguration
}

func (c *confWatcher) Watch(fn func())                      { c.update = fn }
func (c *confWatcher) SiteConfig() schema.SiteConfiguration { return c.conf }

func TestSyncJobsRecordsStoreWatch(t *testing.T) {
	s := newSyncJobsRecordsStore(logtest.Scoped(t))

	// assert default
	assert.IsType(t, noopCache{}, s.cache)

	cw := confWatcher{
		conf: schema.SiteConfiguration{
			AuthzSyncJobsLogsTTL: 5,
		},
	}

	// register
	s.Watch(&cw)

	// proc the update
	cw.update()

	// assert updated
	assert.Equal(t, 5*time.Minute, s.cache.(*rcache.Cache).TTL())
}

type memCache map[string]string

func (m memCache) Set(k string, v []byte) { m[k] = string(v) }

func TestSyncJobRecordsRecord(t *testing.T) {
	s := syncJobsRecordsStore{
		logger: logtest.Scoped(t),
	}
	t.Run("success", func(t *testing.T) {
		c := memCache{}
		s.cache = c
		s.Record("repo", 12, []authz.SyncJobProviderStatus{}, nil)
		autogold.Want("record_success", memCache{"1668537400522946000": `{"request_type":"repo","request_id":12,"completed":"2022-11-15T18:36:40.522946Z","status":"SUCCESS","message":"","providers":[]}`}).
			Equal(t, c)
	})
	t.Run("error", func(t *testing.T) {
		c := memCache{}
		s.cache = c
		s.Record("repo", 12, []authz.SyncJobProviderStatus{}, errors.New("oh no"))
		autogold.Want("record_error", memCache{"1668537400568649000": `{"request_type":"repo","request_id":12,"completed":"2022-11-15T18:36:40.568649Z","status":"ERROR","message":"oh no","providers":[]}`}).
			Equal(t, c)
	})
}
