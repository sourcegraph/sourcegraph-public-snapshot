package syncjobs

import (
	"context"

	"github.com/sourcegraph/sourcegraph/schema"
)

type siteConfigQuerier struct {
	conf schema.SiteConfiguration
}

func (c *siteConfigQuerier) SiteConfig() schema.SiteConfiguration { return c.conf }

type memCache struct {
	// retain in []string for ease of autogold testing
	values []string
}

func (m *memCache) Insert(v []byte) error {
	m.values = append(m.values, string(v))
	return nil
}

func (m *memCache) Slice(ctx context.Context, from, to int) (vals [][]byte, err error) {
	for _, v := range m.values {
		vals = append(vals, []byte(v))
	}
	return
}
