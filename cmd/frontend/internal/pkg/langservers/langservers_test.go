package langservers

import "testing"

func TestStaticInfo_SiteConfig_Language(t *testing.T) {
	// Sanity check that the siteConfig language fields match their map keys,
	// as typos have caused issues here in the past:
	//
	// https://github.com/sourcegraph/sourcegraph/issues/10671
	//
	for lang, staticInfo := range StaticInfo {
		if lang != staticInfo.SiteConfig.Language {
			t.Fatalf("mismatched StaticInfo entry found; lang %q != siteConfig.Language %q", lang, staticInfo.SiteConfig.Language)
		}
	}
}

func TestStaticInfo_debugContainerPorts(t *testing.T) {
	// Sanity check that the languages in StaticInfo and debugContainerPorts are
	// the same.

	for key := range StaticInfo {
		if _, ok := debugContainerPorts[key]; !ok {
			t.Fatalf("debugContainerPorts is missing a key from StaticInfo: %s", key)
		}
	}

	for key := range debugContainerPorts {
		if _, ok := StaticInfo[key]; !ok {
			t.Fatalf("StaticInfo is missing a key from debugContainerPorts: %s", key)
		}
	}
}

func TestDebugContainerPorts_unique(t *testing.T) {
	// Sanity check that all the ports in debugContainerPorts are unique.

	allPorts := make(map[string]string)

	for language, ports := range debugContainerPorts {
		if _, ok := allPorts[ports.HostPort]; ok && language != "javascript" && language != "typescript" {
			t.Fatalf("Languages %s and %s can't both listen on port %s.", language, allPorts[ports.HostPort], ports.HostPort)
		}

		allPorts[ports.HostPort] = language
	}
}
