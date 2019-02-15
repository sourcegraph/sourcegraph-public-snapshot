package conf

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestComputed(t *testing.T) {
	tests := []struct {
		name string
		sc   *Unified
		env  []string
		fun  interface{}
		want interface{}
	}{{
		name: "SearchIndex defaults to false in docker",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=docker-container"},
		fun:  SearchIndexEnabled,
		want: false,
	}, {
		name: "SearchIndex defaults to true in k8s",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=k8s"},
		fun:  SearchIndexEnabled,
		want: true,
	}, {
		name: "SearchIndex enabled",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{SearchIndexEnabled: boolPtr(true)}},
		env:  []string{"DEPLOY_TYPE=docker-container"},
		fun:  SearchIndexEnabled,
		want: true,
	}, {
		name: "SearchIndex disabled",
		sc:   &Unified{SiteConfiguration: schema.SiteConfiguration{SearchIndexEnabled: boolPtr(false)}},
		env:  []string{"DEPLOY_TYPE=docker-container"},
		fun:  SearchIndexEnabled,
		want: false,
	}, {
		name: "SearchIndex INDEXED_SEARCH=f",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=docker-container", "INDEXED_SEARCH=f"},
		fun:  SearchIndexEnabled,
		want: false,
	}, {
		name: "SearchIndex INDEXED_SEARCH=t",
		sc:   &Unified{},
		env:  []string{"DEPLOY_TYPE=docker-container", "INDEXED_SEARCH=t"},
		fun:  SearchIndexEnabled,
		want: true,
	}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, e := range test.env {
				cleanup := setenv(t, e)
				defer cleanup()
			}
			Mock(test.sc)
			got := reflect.ValueOf(test.fun).Call([]reflect.Value{})
			if !reflect.DeepEqual(got[0].Interface(), test.want) {
				t.Fatalf("got %v, want %v", got[0], test.want)
			}
		})
	}
}

func TestDefaultOverride(t *testing.T) {
	tests := []struct {
		name      string
		env       []string
		wantOk    bool
		wantError bool
		check     func(conftypes.RawUnified, bool, error) error
	}{
		name: "DefaultConfigOverride no env vars",
		env:  []string{},
		check: func(sc conftypes.RawUnified, ok bool, err error) error {
			return nil
		},
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
