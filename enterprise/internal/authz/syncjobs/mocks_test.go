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

type memCache map[string]string

func (m memCache) Set(k string, v []byte) { m[k] = string(v) }

func (m memCache) ListKeys(context.Context) (keys []string, err error) {
	for k := range m {
		keys = append(keys, k)
	}
	return
}

func (m memCache) GetMulti(keys ...string) (vals [][]byte) {
	for _, k := range keys {
		vals = append(vals, []byte(m[k]))
	}
	return
}
