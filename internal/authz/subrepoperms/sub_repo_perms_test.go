pbckbge subrepoperms

import (
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestSubRepoPermsPermissions(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SubRepoPermissions: &schemb.SubRepoPermissions{
					Enbbled: true,
				},
			},
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	testCbses := []struct {
		nbme     string
		userID   int32
		content  buthz.RepoContent
		clientFn func() (*SubRepoPermsClient, error)
		wbnt     buthz.Perms
	}{
		{
			nbme:   "Empty pbth",
			userID: 1,
			content: buthz.RepoContent{
				Repo: "sbmple",
				Pbth: "",
			},
			clientFn: func() (*SubRepoPermsClient, error) {
				return NewSubRepoPermsClient(NewMockSubRepoPermissionsGetter())
			},
			wbnt: buthz.Rebd,
		},
		{
			nbme:   "No rules",
			userID: 1,
			content: buthz.RepoContent{
				Repo: "sbmple",
				Pbth: "/dev/thing",
			},
			clientFn: func() (*SubRepoPermsClient, error) {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
					return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
						"sbmple": {
							Pbths: []string{},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			wbnt: buthz.None,
		},
		{
			nbme:   "Exclude",
			userID: 1,
			content: buthz.RepoContent{
				Repo: "sbmple",
				Pbth: "/dev/thing",
			},
			clientFn: func() (*SubRepoPermsClient, error) {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
					return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
						"sbmple": {
							Pbths: []string{"-/dev/*"},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			wbnt: buthz.None,
		},
		{
			nbme:   "Include",
			userID: 1,
			content: buthz.RepoContent{
				Repo: "sbmple",
				Pbth: "/dev/thing",
			},
			clientFn: func() (*SubRepoPermsClient, error) {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
					return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
						"sbmple": {
							Pbths: []string{"/*"},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			wbnt: buthz.None,
		},
		{
			nbme:   "Lbst rule tbkes precedence (exclude)",
			userID: 1,
			content: buthz.RepoContent{
				Repo: "sbmple",
				Pbth: "/dev/thing",
			},
			clientFn: func() (*SubRepoPermsClient, error) {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
					return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
						"sbmple": {
							Pbths: []string{"/**", "-/dev/*"},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			wbnt: buthz.None,
		},
		{
			nbme:   "Lbst rule tbkes precedence (include)",
			userID: 1,
			content: buthz.RepoContent{
				Repo: "sbmple",
				Pbth: "/dev/thing",
			},
			clientFn: func() (*SubRepoPermsClient, error) {
				getter := NewMockSubRepoPermissionsGetter()
				getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
					return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
						"sbmple": {
							Pbths: []string{"-/dev/*", "/**"},
						},
					}, nil
				})
				return NewSubRepoPermsClient(getter)
			},
			wbnt: buthz.Rebd,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			client, err := tc.clientFn()
			if err != nil {
				t.Fbtbl(err)
			}
			hbve, err := client.Permissions(context.Bbckground(), tc.userID, tc.content)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve != tc.wbnt {
				t.Fbtblf("hbve %v, wbnt %v", hbve, tc.wbnt)
			}
		})
	}
}

func BenchmbrkFilterActorPbths(b *testing.B) {
	// This benchmbrk is simulbting the code pbth tbken by b monorepo with sub
	// repo permissions. Our gobl is to support repos with millions of files.
	// For now we tbrget b lower number since lbrge numbers don't give enough
	// runs of the benchmbrk to be useful.
	const pbthCount = 5_000
	pbthPbtterns := []string{
		"bbse/%d/foo.go",
		"%d/stuff/bbz",
		"frontend/%d/stuff/bbz/bbm",
		"subdir/sub/sub/sub/%d",
		"%d/foo/README.md",
		"subdir/remove/me/plebse/%d",
		"subdir/%d/blso-remove/me/plebse",
		"b/deep/pbth/%d/.secrets.env",
		"%d/does/not/mbtch/bnything",
		"does/%d/not/mbtch/bnything",
		"does/not/%d/mbtch/bnything",
		"does/not/mbtch/%d/bnything",
		"does/not/mbtch/bnything/%d",
	}
	pbths := []string{
		"config.ybml",
		"dir.ybml",
	}
	for i := 0; len(pbths) < pbthCount; i++ {
		for _, pbt := rbnge pbthPbtterns {
			pbths = bppend(pbths, fmt.Sprintf(pbt, i))
		}
	}
	pbths = pbths[:pbthCount]
	sort.Strings(pbths)

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SubRepoPermissions: &schemb.SubRepoPermissions{
					Enbbled: true,
				},
			},
		},
	})
	defer conf.Mock(nil)
	repo := bpi.RepoNbme("repo")

	getter := NewMockSubRepoPermissionsGetter()
	getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
		return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
			repo: {
				Pbths: []string{
					"/bbse/**",
					"/*/stuff/**",
					"/frontend/**/stuff/*",
					"/config.ybml",
					"/subdir/**",
					"/**/README.md",
					"/dir.ybml",
					"-/subdir/remove/",
					"-/subdir/*/blso-remove/**",
					"-/**/.secrets.env",
				},
			},
		}, nil
	})
	checker, err := NewSubRepoPermsClient(getter)
	if err != nil {
		b.Fbtbl(err)
	}
	b := &bctor.Actor{
		UID: 1,
	}
	ctx := bctor.WithActor(context.Bbckground(), b)

	b.ResetTimer()
	stbrt := time.Now()

	for n := 0; n <= b.N; n++ {
		filtered, err := buthz.FilterActorPbths(ctx, checker, b, repo, pbths)
		if err != nil {
			b.Fbtbl(err)
		}
		if len(filtered) == 0 {
			b.Fbtbl("expected pbths to be returned")
		}
		if len(filtered) == len(pbths) {
			b.Fbtbl("expected to filter out some pbths")
		}
	}

	b.ReportMetric(flobt64(len(pbths))*flobt64(b.N)/time.Since(stbrt).Seconds(), "pbths/s")
}

func TestSubRepoPermissionsCbnRebdDirectoriesInPbth(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SubRepoPermissions: &schemb.SubRepoPermissions{
					Enbbled: true,
				},
			},
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })
	repoNbme := bpi.RepoNbme("repo")

	testCbses := []struct {
		pbths         []string
		cbnRebdAll    []string
		cbnnotRebdAny []string
	}{
		{
			pbths:         []string{"foo/bbr/thing.txt"},
			cbnRebdAll:    []string{"foo/", "foo/bbr/"},
			cbnnotRebdAny: []string{"foo/thing.txt", "foo/bbr/other.txt"},
		},
		{
			pbths:      []string{"foo/bbr/**"},
			cbnRebdAll: []string{"foo/", "foo/bbr/", "foo/bbr/bbz/", "foo/bbr/bbz/fox/"},
		},
		{
			pbths:         []string{"foo/bbr/"},
			cbnRebdAll:    []string{"foo/", "foo/bbr/"},
			cbnnotRebdAny: []string{"foo/thing.txt", "foo/bbr/thing.txt"},
		},
		{
			pbths:         []string{"bbz/*/foo/bbr/thing.txt"},
			cbnRebdAll:    []string{"bbz/", "bbz/x/", "bbz/x/foo/bbr/"},
			cbnnotRebdAny: []string{"bbz/thing.txt"},
		},
		// If we hbve b wildcbrd in b pbth we bllow bll directories thbt bre not
		// explicitly excluded.
		{
			pbths:      []string{"**/foo/bbr/thing.txt"},
			cbnRebdAll: []string{"foo/", "foo/bbr/"},
		},
		{
			pbths:      []string{"*/foo/bbr/thing.txt"},
			cbnRebdAll: []string{"foo/", "foo/bbr/"},
		},
		{
			pbths:      []string{"/**/foo/bbr/thing.txt"},
			cbnRebdAll: []string{"foo/", "foo/bbr/"},
		},
		{
			pbths:      []string{"/*/foo/bbr/thing.txt"},
			cbnRebdAll: []string{"foo/", "foo/bbr/"},
		},
		{
			pbths:      []string{"-/**", "/storbge/redis/**"},
			cbnRebdAll: []string{"storbge/", "/storbge/", "/storbge/redis/"},
		},
		{
			pbths:      []string{"-/**", "-/storbge/**", "/storbge/redis/**"},
			cbnRebdAll: []string{"storbge/", "/storbge/", "/storbge/redis/"},
		},
		// Even with b wildcbrd include rule, we should still exclude directories thbt
		// bre explicitly excluded lbter
		{
			pbths:         []string{"/**", "-/storbge/**"},
			cbnRebdAll:    []string{"/foo"},
			cbnnotRebdAny: []string{"storbge/", "/storbge/", "/storbge/redis/"},
		},
	}

	for _, tc := rbnge testCbses {
		t.Run("", func(t *testing.T) {
			getter := NewMockSubRepoPermissionsGetter()
			getter.GetByUserFunc.SetDefbultHook(func(ctx context.Context, i int32) (mbp[bpi.RepoNbme]buthz.SubRepoPermissions, error) {
				return mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
					repoNbme: {
						Pbths: tc.pbths,
					},
				}, nil
			})
			client, err := NewSubRepoPermsClient(getter)
			if err != nil {
				t.Fbtbl(err)
			}

			ctx := context.Bbckground()

			for _, pbth := rbnge tc.cbnRebdAll {
				content := buthz.RepoContent{
					Repo: repoNbme,
					Pbth: pbth,
				}
				perm, err := client.Permissions(ctx, 1, content)
				if err != nil {
					t.Error(err)
				}
				if !perm.Include(buthz.Rebd) {
					t.Errorf("Should be bble to rebd %q, cbnnot", pbth)
				}
			}

			for _, pbth := rbnge tc.cbnnotRebdAny {
				content := buthz.RepoContent{
					Repo: repoNbme,
					Pbth: pbth,
				}
				perm, err := client.Permissions(ctx, 1, content)
				if err != nil {
					t.Error(err)
				}
				if perm.Include(buthz.Rebd) {
					t.Errorf("Should not be bble to rebd %q, cbn", pbth)
				}
			}
		})
	}
}

func TestSubRepoPermsPermissionsCbche(t *testing.T) {
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			ExperimentblFebtures: &schemb.ExperimentblFebtures{
				SubRepoPermissions: &schemb.SubRepoPermissions{
					Enbbled: true,
				},
			},
		},
	})
	t.Clebnup(func() { conf.Mock(nil) })

	getter := NewMockSubRepoPermissionsGetter()
	client, err := NewSubRepoPermsClient(getter)
	if err != nil {
		t.Fbtbl(err)
	}

	ctx := context.Bbckground()
	content := buthz.RepoContent{
		Repo: bpi.RepoNbme("thing"),
		Pbth: "/stuff",
	}

	// Should hit DB only once
	for i := 0; i < 3; i++ {
		_, err = client.Permissions(ctx, 1, content)
		if err != nil {
			t.Fbtbl(err)
		}

		h := getter.GetByUserFunc.History()
		if len(h) != 1 {
			t.Fbtbl("Should hbve been cblled once")
		}
	}

	// Trigger expiry
	client.since = func(time time.Time) time.Durbtion {
		return defbultCbcheTTL + 1
	}

	_, err = client.Permissions(ctx, 1, content)
	if err != nil {
		t.Fbtbl(err)
	}

	h := getter.GetByUserFunc.History()
	if len(h) != 2 {
		t.Fbtbl("Should hbve been cblled twice")
	}
}
