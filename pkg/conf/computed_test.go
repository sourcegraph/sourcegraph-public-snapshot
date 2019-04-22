package conf

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/pkg/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"os"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestSearchIndexEnabled(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		env  []string
		want interface{}
	}{{
		name: "SearchIndex defaults to false in docker",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=docker-container"},
		want: false,
	}, {
		name: "SearchIndex defaults to true in k8s",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=k8s"},
		want: true,
	}, {
		name: "SearchIndex enabled",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{SearchIndexEnabled: boolPtr(true)}},
		env:  []string{"DEPLOY_TYPE=docker-container"},
		want: true,
	}, {
		name: "SearchIndex disabled",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{SearchIndexEnabled: boolPtr(false)}},
		env:  []string{"DEPLOY_TYPE=docker-container"},
		want: false,
	}, {
		name: "SearchIndex INDEXED_SEARCH=f",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=docker-container", "INDEXED_SEARCH=f"},
		want: false,
	}, {
		name: "SearchIndex INDEXED_SEARCH=t",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=docker-container", "INDEXED_SEARCH=t"},
		want: true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, e := range test.env {
				cleanup := setenv(t, e)
				defer cleanup()
			}
			Mock(test.sc)
			got := SearchIndexEnabled()
			if got != test.want {
				t.Fatalf("SearchIndexEnabled() = %v, want %v", got, test.want)
			}
		})
	}

	defaults := map[string]conftypes.RawUnified{
		"Cluster": confdefaults.Cluster,
		"Default": confdefaults.Default,
		"DevAndTesting": confdefaults.DevAndTesting,
		"DockerContainer": confdefaults.DockerContainer,
	}
	for dStr, d := range defaults {
		test := fmt.Sprintf("for %s defaults", dStr)
		t.Run(test, func(t *testing.T) {
			cfg, err := ParseConfig(d)
			if err != nil {
				t.Fatal(err)
			}
			Mock(cfg)
			defer Mock(nil)
			if !SearchIndexEnabled() {
				t.Errorf("search indexing should be enabled by default for Docker deployments")
			}
		})
	}
}

func setenv(t *testing.T, keyval string) func() {
	t.Helper()

	parts := strings.SplitN(keyval, "=", 2)
	key := parts[0]
	value := parts[1]

	orig, set := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatal(err)
	}
	if set {
		return func() {
			if err := os.Setenv(key, orig); err != nil {
				t.Fatal(err)
			}
		}
	}
	return func() {
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

func boolPtr(b bool) *bool {
	return &b
}
