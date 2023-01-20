package syncjobs

import (
	"context"

	"github.com/sourcegraph/sourcegraph/schema"
)

type confWatcher struct {
	update func()
	conf   schema.SiteConfiguration
}

func (c *confWatcher) Watch(fn func())                      { c.update = fn }
func (c *confWatcher) SiteConfig() schema.SiteConfiguration { return c.conf }

type memCache struct {
	// retain in []string for ease of autogold testing
	values []string
}

func (m *memCache) Insert(v []byte) error {
	m.values = append(m.values, string(v))
	return nil
}

// no-op
func (m *memCache) SetMaxSize(int) {}

func (m *memCache) Slice(ctx context.Context, from, to int) (vals [][]byte, err error) {
	for _, v := range m.values {
		vals = append(vals, []byte(v))
	}
	return
}
