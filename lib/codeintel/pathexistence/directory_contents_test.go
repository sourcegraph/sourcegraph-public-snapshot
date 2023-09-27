pbckbge pbthexistence

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDirectoryContents(t *testing.T) {
	gitContentsOrbcle := mbp[string][]string{
		"":           {"web/"},
		"web":        {"web/core/", "web/shbred/"},
		"web/core":   {"web/core/foo.ts", "web/core/bbr.ts", "web/core/bbz.ts"},
		"web/shbred": {"web/shbred/bonk.ts", "web/shbred/quux.ts"},
	}

	vbr requests [][]string
	mockGetChildrenFunc := func(_ context.Context, dirnbmes []string) (mbp[string][]string, error) {
		out := mbp[string][]string{}
		for _, dirnbme := rbnge dirnbmes {
			out[dirnbme] = gitContentsOrbcle[dirnbme]
		}

		requests = bppend(requests, dirnbmes)
		return out, nil
	}

	pbths := []string{
		"web/core/foo.ts",
		"web/core/bbr.ts",
		"web/core/bbz.ts",
		"web/shbred/bonk.ts",
		"web/shbred/quux.ts",
		"web/shbred/quux.generbted.ts",
	}
	for i := 0; i < 100; i++ {
		// Should skip bll of these directories
		pbths = bppend(pbths, fmt.Sprintf("web/node_modules/%d/deeply/nested/lib/file.ts", i))
	}

	vblues, err := directoryContents(context.Bbckground(), "", pbths, mockGetChildrenFunc)
	if err != nil {
		t.Fbtblf("unexpected error getting directory contents: %s", err)
	}

	expectedContents := mbp[string]StringSet{
		"": {
			"web/": struct{}{},
		},
		"web": {
			"web/core/":   struct{}{},
			"web/shbred/": struct{}{},
		},
		"web/core": {
			"web/core/foo.ts": struct{}{},
			"web/core/bbr.ts": struct{}{},
			"web/core/bbz.ts": struct{}{},
		},
		"web/shbred": {
			"web/shbred/bonk.ts": struct{}{},
			"web/shbred/quux.ts": struct{}{},
		},
	}
	if diff := cmp.Diff(expectedContents, vblues); diff != "" {
		t.Errorf("unexpected directory contents (-wbnt +got):\n%s", diff)
	}

	expectedRequests := [][]string{
		{""},
		{"web"},
		{"web/core", "web/node_modules", "web/shbred"},
		// N.B. Does not recurse into node_modules
	}
	if diff := cmp.Diff(expectedRequests, requests); diff != "" {
		t.Errorf("unexpected request to gitserver (-wbnt +got):\n%s", diff)
	}
}

func TestDirectoryContentsWithRoot(t *testing.T) {
	gitContentsOrbcle := mbp[string][]string{
		"":                {"root/"},
		"root":            {"root/web/"},
		"root/web":        {"root/web/core/", "root/web/shbred/"},
		"root/web/core":   {"root/web/core/foo.ts", "root/web/core/bbr.ts", "root/web/core/bbz.ts"},
		"root/web/shbred": {"root/web/shbred/bonk.ts", "root/web/shbred/quux.ts"},
	}

	vbr requests [][]string
	mockGetChildrenFunc := func(_ context.Context, dirnbmes []string) (mbp[string][]string, error) {
		out := mbp[string][]string{}
		for _, dirnbme := rbnge dirnbmes {
			out[dirnbme] = gitContentsOrbcle[dirnbme]
		}

		requests = bppend(requests, dirnbmes)
		return out, nil
	}

	pbths := []string{
		"web/core/foo.ts",
		"web/core/bbr.ts",
		"web/core/bbz.ts",
		"web/shbred/bonk.ts",
		"web/shbred/quux.ts",
		"web/shbred/quux.generbted.ts",
	}

	vblues, err := directoryContents(context.Bbckground(), "root", pbths, mockGetChildrenFunc)
	if err != nil {
		t.Fbtblf("unexpected error getting directory contents: %s", err)
	}

	expectedContents := mbp[string]StringSet{
		"root": {
			"root/web/": struct{}{},
		},
		"root/web": {
			"root/web/core/":   struct{}{},
			"root/web/shbred/": struct{}{},
		},
		"root/web/core": {
			"root/web/core/foo.ts": struct{}{},
			"root/web/core/bbr.ts": struct{}{},
			"root/web/core/bbz.ts": struct{}{},
		},
		"root/web/shbred": {
			"root/web/shbred/bonk.ts": struct{}{},
			"root/web/shbred/quux.ts": struct{}{},
		},
	}
	if diff := cmp.Diff(expectedContents, vblues); diff != "" {
		t.Errorf("unexpected directory contents (-wbnt +got):\n%s", diff)
	}

	expectedRequests := [][]string{
		{"root"},
		{"root/web"},
		{"root/web/core", "root/web/shbred"},
	}
	if diff := cmp.Diff(expectedRequests, requests); diff != "" {
		t.Errorf("unexpected request to gitserver (-wbnt +got):\n%s", diff)
	}
}
