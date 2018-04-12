package langservers

import "testing"

func TestStaticInfo_SiteConfig_Language(t *testing.T) {
	// Sanity check that the siteConfig language fields match their map keys,
	// as typos have caused issues here in the past:
	//
	// https://github.com/sourcegraph/sourcegraph/issues/10671
	//
	for lang, staticInfo := range StaticInfo {
		if lang != staticInfo.siteConfig.Language {
			t.Fatalf("mismatched StaticInfo entry found; lang %q != siteConfig.Language %q", lang, staticInfo.siteConfig.Language)
		}
	}
}
