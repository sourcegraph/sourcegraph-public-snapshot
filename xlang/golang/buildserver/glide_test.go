package buildserver

import "testing"

func TestLoadGlideLock(t *testing.T) {
	yml := []byte(`hash: 8aeb29a35adb31f8b46a792ae1b304c2c55f2d10bfe0ca1a4b8ac5330e22decc
updated: 2016-11-09T16:14:48.657534669+09:00
imports:
- name: github.com/cactus/go-statsd-client
  version: d8eabe07bc70ff9ba6a56836cde99d1ea3d005f7
  subpackages:
  - statsd
- name: github.com/Sirupsen/logrus
  version: 1445b7a38228c041834afc69231b7966b9943397
- name: github.com/uber-common/bark
  version: 8841a0f8e7ca869284ccb29c08a14cf3f4310f46
- name: github.com/uber-go/atomic
  version: 9e99152552a6ce13fa3b2ce4a9c4fb117cca4506
- name: golang.org/x/sys
  version: 9a2e24c3733eddc63871eda99f253e2db29bd3b9
  subpackages:
  - unix
testImports:
- name: github.com/apex/log
  version: 4ea85e918cc8389903d5f12d7ccac5c23ab7d89b
  subpackages:
  - handlers/json
`)
	cases := map[string]string{
		// Specified in yaml
		"github.com/cactus/go-statsd-client":        "d8eabe07bc70ff9ba6a56836cde99d1ea3d005f7",
		"github.com/cactus/go-statsd-client/statsd": "d8eabe07bc70ff9ba6a56836cde99d1ea3d005f7",
		"github.com/Sirupsen/logrus":                "1445b7a38228c041834afc69231b7966b9943397",
		"github.com/uber-common/bark":               "8841a0f8e7ca869284ccb29c08a14cf3f4310f46",
		"github.com/uber-go/atomic":                 "9e99152552a6ce13fa3b2ce4a9c4fb117cca4506",
		"golang.org/x/sys":                          "9a2e24c3733eddc63871eda99f253e2db29bd3b9",
		"golang.org/x/sys/unix":                     "9a2e24c3733eddc63871eda99f253e2db29bd3b9",
		"github.com/apex/log":                       "4ea85e918cc8389903d5f12d7ccac5c23ab7d89b",
		"github.com/apex/log/handlers/json":         "4ea85e918cc8389903d5f12d7ccac5c23ab7d89b",
		"github.com/apex/log/handlers/logfmt":       "4ea85e918cc8389903d5f12d7ccac5c23ab7d89b",

		// Not specified
		"github/a":          "",
		"github/a/a":        "",
		"github/a/a/a":      "",
		"golang.org/x/syss": "",
		"golang.org/x/sy":   "",
		"golang.org/x/sy/s": "",
		"z.com/z/z":         "",
		"fmt":               "",
	}
	p := loadGlideLock(yml)
	for pkg, want := range cases {
		got := p.Find(pkg)
		if got != want {
			t.Errorf("Find(%v) = %v, want %v", pkg, got, want)
		}
	}
}
