package conf

import (
	"os"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestExtensions(t *testing.T) {
	check := func(t *testing.T, want *PlatformConfiguration) {
		t.Helper()
		got := Extensions()
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	}

	t.Run("no config and no DefaultRemoteRegistry", func(t *testing.T) {
		DefaultRemoteRegistry = ""
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{}})
		check(t, nil)
	})

	t.Run("no config but valid DefaultRemoteRegistry", func(t *testing.T) {
		DefaultRemoteRegistry = "x"
		defer func() { DefaultRemoteRegistry = "" }()
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{}})
		check(t, &PlatformConfiguration{RemoteRegistryURL: "x"})
	})

	t.Run("empty config, valid DefaultRemoteRegistry", func(t *testing.T) {
		DefaultRemoteRegistry = "x"
		defer func() { DefaultRemoteRegistry = "" }()
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{}}})
		check(t, &PlatformConfiguration{RemoteRegistryURL: "x"})
	})

	t.Run("config extensions.disabled", func(t *testing.T) {
		DefaultRemoteRegistry = "x"
		defer func() { DefaultRemoteRegistry = "" }()
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{Disabled: boolPtr(true)}}})
		check(t, nil)
	})

	t.Run("config extensions.disabled false", func(t *testing.T) {
		DefaultRemoteRegistry = "x"
		defer func() { DefaultRemoteRegistry = "" }()
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{Disabled: boolPtr(false)}}})
		check(t, &PlatformConfiguration{RemoteRegistryURL: "x"})
	})

	t.Run("config extensions.remoteRegistry overrides DefaultRemoteRegistry", func(t *testing.T) {
		DefaultRemoteRegistry = "x"
		defer func() { DefaultRemoteRegistry = "" }()
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{RemoteRegistry: "y"}}})
		check(t, &PlatformConfiguration{RemoteRegistryURL: "y"})
	})

	t.Run("config extensions.remoteRegistry false", func(t *testing.T) {
		DefaultRemoteRegistry = "x"
		defer func() { DefaultRemoteRegistry = "" }()
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{Extensions: &schema.Extensions{RemoteRegistry: false}}})
		check(t, &PlatformConfiguration{RemoteRegistryURL: ""})
	})

	t.Run("OFFLINE env var", func(t *testing.T) {
		os.Setenv("OFFLINE", "1")
		defer os.Unsetenv("OFFLINE")
		DefaultRemoteRegistry = "x"
		defer func() { DefaultRemoteRegistry = "" }()
		Mock(&Unified{SiteConfiguration: schema.SiteConfiguration{}})
		check(t, &PlatformConfiguration{RemoteRegistryURL: ""})
	})
}
