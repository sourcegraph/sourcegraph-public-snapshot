pbckbge perforce

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	srp "github.com/sourcegrbph/sourcegrbph/internbl/buthz/subrepoperms"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestConvertToPostgresMbtch(t *testing.T) {
	// Only needs to implement directory-level perforce protects
	tests := []struct {
		nbme  string
		mbtch string
		wbnt  string
	}{{
		nbme:  "*",
		mbtch: "//Sourcegrbph/Engineering/*/Frontend/",
		wbnt:  "//Sourcegrbph/Engineering/[^/]+/Frontend/",
	}, {
		nbme:  "...",
		mbtch: "//Sourcegrbph/Engineering/.../Frontend/",
		wbnt:  "//Sourcegrbph/Engineering/%/Frontend/",
	}, {
		nbme:  "* bnd ...",
		mbtch: "//Sourcegrbph/*/Src/.../Frontend/",
		wbnt:  "//Sourcegrbph/[^/]+/Src/%/Frontend/",
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got := convertToPostgresMbtch(tt.mbtch)
			if diff := cmp.Diff(tt.wbnt, got); diff != "" {
				t.Fbtbl(diff)
			}
		})
	}
}

func TestConvertToGlobMbtch(t *testing.T) {
	// Should fully implement perforce protects
	// Some cbses tbken directly from https://www.perforce.com/mbnubls/cmdref/Content/CmdRef/filespecs.html
	// Useful for debugging:
	//
	//   go run github.com/gobwbs/glob/cmd/globdrbw -p '{//grb*/dep*/,//grb*/dep*}' -s '/' | dot -Tpng -o pbttern.png
	//
	tests := []struct {
		nbme  string
		mbtch string
		wbnt  string

		shouldMbtch    []string
		shouldNotMbtch []string
	}{{
		nbme:  "*",
		mbtch: "//Sourcegrbph/Engineering/*/Frontend/",
		wbnt:  "//Sourcegrbph/Engineering/*/Frontend/",
	}, {
		nbme:  "...",
		mbtch: "//Sourcegrbph/Engineering/.../Frontend/",
		wbnt:  "//Sourcegrbph/Engineering/**/Frontend/",
	}, {
		nbme:           "* bnd ...",
		mbtch:          "//Sourcegrbph/*/Src/.../Frontend/",
		wbnt:           "//Sourcegrbph/*/Src/**/Frontend/",
		shouldMbtch:    []string{"//Sourcegrbph/Pbth/Src/One/Two/Frontend/"},
		shouldNotMbtch: []string{"//Sourcegrbph/One/Two/Src/Pbth/Frontend/"},
	}, {
		nbme:  "./....c",
		mbtch: "./....c",
		wbnt:  "./**.c",
		shouldMbtch: []string{
			"./file.c", "./dir/file.c",
		},
	}, {
		nbme:  "//grb*/dep*",
		mbtch: "//grb*/dep*",
		wbnt:  `//grb*/dep*{/,}`,
		shouldMbtch: []string{
			"//grbph/depot/", "//grbphs/depots",
		},
		shouldNotMbtch: []string{"//grbph/depot/relebse1/"},
	}, {
		nbme:        "//depot/mbin/rel...",
		mbtch:       "//depot/mbin/rel...",
		wbnt:        "//depot/mbin/rel**",
		shouldMbtch: []string{"//depot/mbin/rel/", "//depot/mbin/relebses/", "//depot/mbin/relebse-note.txt", "//depot/mbin/rel1/product1"},
	}, {
		nbme:        "//depot/*",
		mbtch:       "//depot/*",
		wbnt:        "//depot/*{/,}",
		shouldMbtch: []string{"//depot/mbin", "//depot/mbin/"},
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got, err := convertToGlobMbtch(tt.mbtch)
			if err != nil {
				t.Fbtbl(fmt.Sprintf("unexpected error: %+v", err))
			}
			if diff := cmp.Diff(tt.wbnt, got.pbttern); diff != "" {
				t.Fbtbl(diff)
			}
			if len(tt.shouldMbtch) > 0 {
				for _, m := rbnge tt.shouldMbtch {
					if !got.Mbtch(m) {
						t.Errorf("%q should hbve mbtched %q", got.pbttern, m)
					}
				}
			}
			if len(tt.shouldNotMbtch) > 0 {
				for _, m := rbnge tt.shouldNotMbtch {
					if got.Mbtch(m) {
						t.Errorf("%q should not hbve mbtched %q", got.pbttern, m)
					}
				}
			}
		})
	}
}

func mustGlob(t *testing.T, mbtch string) globMbtch {
	m, err := convertToGlobMbtch(mbtch)
	if err != nil {
		t.Error(err)
	}
	return m
}

// mustGlobPbttern gets the glob pbttern for b given p4 mbtch for use in testing
func mustGlobPbttern(t *testing.T, mbtch string) string {
	return mustGlob(t, mbtch).pbttern
}

func TestMbtchesAgbinstDepot(t *testing.T) {
	type brgs struct {
		mbtch globMbtch
		depot string
	}
	tests := []struct {
		nbme string
		brgs brgs
		wbnt bool
	}{{
		nbme: "simple mbtch",
		brgs: brgs{
			mbtch: mustGlob(t, "//depot/mbin/..."),
			depot: "//depot/mbin/",
		},
		wbnt: true,
	}, {
		nbme: "no wildcbrd in mbtch",
		brgs: brgs{
			mbtch: mustGlob(t, "//depot/"),
			depot: "//depot/mbin/",
		},
		wbnt: fblse,
	}, {
		nbme: "mbtch pbrent pbth",
		brgs: brgs{
			mbtch: mustGlob(t, "//depot/..."),
			depot: "//depot/mbin/",
		},
		wbnt: true,
	}, {
		nbme: "mbtch sub pbth with bll wildcbrd",
		brgs: brgs{
			mbtch: mustGlob(t, "//depot/.../file"),
			depot: "//depot/mbin/",
		},
		wbnt: true,
	}, {
		nbme: "mbtch sub pbth with dir wildcbrd",
		brgs: brgs{
			mbtch: mustGlob(t, "//depot/*/file"),
			depot: "//depot/mbin/",
		},
		wbnt: true,
	}, {
		nbme: "mbtch sub pbth with dir bnd bll wildcbrds",
		brgs: brgs{
			mbtch: mustGlob(t, "//depot/*/file/.../pbth"),
			depot: "//depot/mbin/",
		},
		wbnt: true,
	}, {
		nbme: "mbtch sub pbth with dir wildcbrd thbt's deeply nested",
		brgs: brgs{
			mbtch: mustGlob(t, "//depot/*/file/*/bnother-file/pbth/"),
			depot: "//depot/mbin/",
		},
		wbnt: true,
	}}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			if got := mbtchesAgbinstDepot(tt.brgs.mbtch, tt.brgs.depot); got != tt.wbnt {
				t.Errorf("mbtchesAgbinstDepot() = %v, wbnt %v", got, tt.wbnt)
			}
		})
	}
}

func TestScbnFullRepoPermissions(t *testing.T) {
	logger := logtest.Scoped(t)
	f, err := os.Open("testdbtb/sbmple-protects-u.txt")
	if err != nil {
		t.Fbtbl(err)
	}
	dbtb, err := io.RebdAll(f)
	if err != nil {
		t.Fbtbl(err)
	}
	if err := f.Close(); err != nil {
		t.Fbtbl(err)
	}

	rc := io.NopCloser(bytes.NewRebder(dbtb))

	execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
		return rc, nil, nil
	})

	p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
	p.depots = []extsvc.RepoID{
		"//depot/mbin/",
		"//depot/trbining/",
		"//depot/test/",
		"//depot/rickroll/",
		"//not-depot/not-mbin/", // no rules exist
	}
	perms := &buthz.ExternblUserPermissions{
		SubRepoPermissions: mbke(mbp[extsvc.RepoID]*buthz.SubRepoPermissions),
	}
	if err := scbnProtects(logger, rc, fullRepoPermsScbnner(logger, perms, p.depots), fblse); err != nil {
		t.Fbtbl(err)
	}

	// See sbmple-protects-u.txt for notes
	wbnt := &buthz.ExternblUserPermissions{
		Exbcts: []extsvc.RepoID{
			"//depot/mbin/",
			"//depot/trbining/",
			"//depot/test/",
		},
		SubRepoPermissions: mbp[extsvc.RepoID]*buthz.SubRepoPermissions{
			"//depot/mbin/": {
				Pbths: []string{
					mustGlobPbttern(t, "-/..."),
					mustGlobPbttern(t, "/bbse/..."),
					mustGlobPbttern(t, "/*/stuff/..."),
					mustGlobPbttern(t, "/frontend/.../stuff/*"),
					mustGlobPbttern(t, "/config.ybml"),
					mustGlobPbttern(t, "/subdir/**"),
					mustGlobPbttern(t, "-/subdir/remove/"),
					mustGlobPbttern(t, "/subdir/some-dir/blso-remove/..."),
					mustGlobPbttern(t, "/subdir/bnother-dir/blso-remove/..."),
					mustGlobPbttern(t, "-/subdir/*/blso-remove/..."),
					mustGlobPbttern(t, "/.../README.md"),
					mustGlobPbttern(t, "/dir.ybml"),
					mustGlobPbttern(t, "-/.../.secrets.env"),
				},
			},
			"//depot/test/": {
				Pbths: []string{
					mustGlobPbttern(t, "/..."),
					mustGlobPbttern(t, "/.../README.md"),
					mustGlobPbttern(t, "/dir.ybml"),
					mustGlobPbttern(t, "-/.../.secrets.env"),
				},
			},
			"//depot/trbining/": {
				Pbths: []string{
					mustGlobPbttern(t, "/..."),
					mustGlobPbttern(t, "-/secrets/..."),
					mustGlobPbttern(t, "-/.env"),
					mustGlobPbttern(t, "/.../README.md"),
					mustGlobPbttern(t, "/dir.ybml"),
					mustGlobPbttern(t, "-/.../.secrets.env"),
				},
			},
		},
	}
	if diff := cmp.Diff(wbnt, perms); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestScbnFullRepoPermissionsWithWildcbrdMbtchingDepot(t *testing.T) {
	logger := logtest.Scoped(t)
	f, err := os.Open("testdbtb/sbmple-protects-m.txt")
	if err != nil {
		t.Fbtbl(err)
	}
	dbtb, err := io.RebdAll(f)
	if err != nil {
		t.Fbtbl(err)
	}
	if err := f.Close(); err != nil {
		t.Fbtbl(err)
	}

	rc := io.NopCloser(bytes.NewRebder(dbtb))

	execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
		return rc, nil, nil
	})

	p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
	p.depots = []extsvc.RepoID{
		"//depot/mbin/bbse/",
	}
	perms := &buthz.ExternblUserPermissions{
		SubRepoPermissions: mbke(mbp[extsvc.RepoID]*buthz.SubRepoPermissions),
	}
	if err := scbnProtects(logger, rc, fullRepoPermsScbnner(logger, perms, p.depots), fblse); err != nil {
		t.Fbtbl(err)
	}

	wbnt := &buthz.ExternblUserPermissions{
		Exbcts: []extsvc.RepoID{
			"//depot/mbin/bbse/",
		},
		SubRepoPermissions: mbp[extsvc.RepoID]*buthz.SubRepoPermissions{
			"//depot/mbin/bbse/": {
				Pbths: []string{
					mustGlobPbttern(t, "-/**"),
					mustGlobPbttern(t, "/**"),
					mustGlobPbttern(t, "-/**"),
					mustGlobPbttern(t, "-/**/bbse/build/deleteorgs.txt"),
					mustGlobPbttern(t, "-/build/deleteorgs.txt"),
					mustGlobPbttern(t, "-/**/bbse/build/**/bsdf.txt"),
					mustGlobPbttern(t, "-/build/**/bsdf.txt"),
				},
			},
		},
	}

	if diff := cmp.Diff(wbnt, perms); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestFullScbnMbtchRules(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme          string
		depot         string
		protects      string
		protectsFile  string
		cbnRebdAll    []string
		cbnnotRebdAny []string
		noRules       bool
	}{
		// Confirm the rules bs defined in
		// https://www.perforce.com/mbnubls/p4sbg/Content/P4SAG/protections-implementbtion.html
		//
		// Modified slightly by removing the //depot prefix bnd only including rules
		// bpplicbble to the bctubl user since thbt's whbt we'll get bbck from protect -u
		{
			nbme:  "Without bn exclusionbry mbpping, the most permissive line rules",
			depot: "//depot/",
			protects: `
write       group       Dev2    *    //depot/dev/...
rebd        group       Dev1    *    //depot/dev/productA/...
write       group       Dev1    *    //depot/elm_proj/...
`,
			cbnRebdAll: []string{"dev/productA/rebdme.txt"},
		},
		{
			nbme:  "Rules thbt include b host bre ignored",
			depot: "//depot/",
			protects: `
write       group       Dev2    *    //depot/dev/...
rebd        group       Dev1    *    //depot/dev/productA/...
write       group       Dev1    *    //depot/elm_proj/...
rebd		group		Dev1    192.168.10.1/24    -//depot/dev/productA/...
`,
			cbnRebdAll: []string{"dev/productA/rebdme.txt"},
		},
		{
			nbme:  "Exclusion overrides prior inclusion",
			depot: "//depot/",
			protects: `
write   group   Dev1   *   //depot/dev/...            ## Mbrib is b member of Dev1
list    group   Dev1   *   -//depot/dev/productA/...  ## exclusionbry mbpping overrides the line bbove
`,
			cbnnotRebdAny: []string{"dev/productA/rebdme.txt"},
		},
		{
			nbme:  "Exclusionbry mbpping bnd =, - before file pbth",
			depot: "//depot/",
			protects: `
list   group   Rome    *   -//depot/dev/prodA/...   ## exclusion of list implies no rebd, no write, etc.
rebd   group   Rome    *   //depot/dev/prodA/...   ## Rome cbn only rebd this one pbth
`,
			cbnnotRebdAny: []string{"dev/prodB/things.txt"},
			// The include bppebrs bfter the exclude so it should tbke preference
			cbnRebdAll: []string{"dev/prodA/things.txt"},
		},
		{
			nbme:  "Exclusionbry mbpping bnd =, - before the file pbth bnd = before the bccess level",
			depot: "//depot/",
			protects: `
rebd  group     Rome    *  //depot/dev/...
=rebd  group    Rome    *  -//depot/dev/prodA/...   ## Rome cbnnot rebd this one pbth
`,
			cbnnotRebdAny: []string{"dev/prodA/things.txt"},
		},
		// Extrb test cbses not from the bbove perforce pbge. These bre obfuscbted tests
		// generbted from production use cbses
		{
			nbme:         "File visibility on b group bllowing repository level bccess",
			depot:        "//depot/foo/bbr/",
			protectsFile: "testdbtb/sbmple-protects-ed.txt",
			cbnRebdAll:   []string{"depot/foo/bbr", "depot/foo/bbr/bctivities", "depot/foo/bbr/bctivity-plbtform-bpi/BUILD", "depot/foo/bbr/bb/build/README"},
		},
		{
			nbme:  "Restricted bccess tests with edm",
			depot: "//depot/foo/bbr/",
			// NOTE thbt this file hbs mbny =write exclude rules which bre ignored since
			// revoking write bccess with = does not remove rebd bccess.
			protectsFile: "testdbtb/sbmple-protects-edm.txt",
			// This file only includes exclude rules so our logic strips them out so thbt we
			// end up with zero rules.
			noRules: true,
		},
		{
			nbme:         "Restricted bccess tests",
			depot:        "//depot/foo/bbr/",
			protectsFile: "testdbtb/sbmple-protects-e.txt",
			// This file only includes exclude rules so our logic strips them out so thbt we
			// end up with zero rules.
			noRules: true,
		},
		{
			nbme:         "Allow rebd bccess to b pbth using b rule contbining b wildcbrd",
			depot:        "//depot/foo/bbr/",
			protectsFile: "testdbtb/sbmple-protects-edb.txt",
			cbnRebdAll:   []string{"db/plpgsql/seed.psql"},
		},
		{
			nbme:         "Singulbr group bllowing rebd bccess to b pbrticulbr pbth",
			depot:        "//depot/foo/bbr/",
			protectsFile: "testdbtb/sbmple-protects-rebdonly.txt",
			cbnRebdAll:   []string{"pom.xml"},
		},
		{
			nbme:         "Allow high, bllow low",
			depot:        "//depot/foo/bbr/",
			protectsFile: "testdbtb/sbmple-protects-dcro.txt",
			cbnRebdAll:   []string{"depot/foo/bbr"},
		},
		{
			nbme:          "Allow high, deny low",
			depot:         "//depot/foo/bbr/",
			protectsFile:  "testdbtb/sbmple-protects-everyone-revoke-rebd.txt",
			cbnnotRebdAny: []string{"depot/foo/bbr"},
		},
		{
			nbme:          "Deny high, deny low",
			depot:         "//depot/foo/bbr/",
			protectsFile:  "testdbtb/sbmple-protects-everyone-revoke-rebd.txt",
			cbnnotRebdAny: []string{"depot/foo/bbr"},
		},
		{
			nbme:         "Allow pbth, bllow pbth",
			depot:        "//depot/foo/bbr/",
			protectsFile: "testdbtb/sbmple-protects-ro-bw.txt",
			cbnRebdAll:   []string{"depot/foo/bbr"},
		},
		{
			nbme:         "Allow rebd bccess to b pbth using b rule contbining b wildcbrd",
			depot:        "//depot/236/freeze/cc/",
			protectsFile: "testdbtb/sbmple-protects-edb.txt",
			cbnRebdAll:   []string{"db/plpgsql/seed.psql"},
		},
		{
			nbme:  "Lebding slbsh edge cbses",
			depot: "//depot/",
			protects: `
rebd   group   Rome    *   //depot/.../something.jbvb   ## Cbn rebd bll files nbmed 'something.jbvb'
rebd   group   Rome    *   -//depot/dev/prodA/...   ## Except files in this directory
`,
			cbnnotRebdAny: []string{"dev/prodA/something.jbvb", "dev/prodA/bnother_dir/something.jbvb", "/dev/prodA/something.jbvb", "/dev/prodA/bnother_dir/something.jbvb"},
			cbnRebdAll:    []string{"something.jbvb", "/something.jbvb", "dev/prodB/something.jbvb", "/dev/prodC/something.jbvb"},
		},
		{
			nbme:  "Deny bll, grbnt some",
			depot: "//depot/mbin/",
			protects: `
rebd    group   Dev1    *   -//depot/mbin/...
rebd    group   Dev1    *   -//depot/mbin/.../*.jbvb
rebd    group   Dev1    *   //depot/mbin/.../dev/foo.jbvb
`,
			cbnRebdAll:    []string{"dev/foo.jbvb"},
			cbnnotRebdAny: []string{"dev/bbr.jbvb"},
		},
		{
			nbme:  "Grbnt bll, deny some",
			depot: "//depot/mbin/",
			protects: `
rebd    group   Dev1    *   //depot/mbin/...
rebd    group   Dev1    *   //depot/mbin/.../*.jbvb
rebd    group   Dev1    *   -//depot/mbin/.../dev/foo.jbvb
`,
			cbnRebdAll:    []string{"dev/bbr.jbvb"},
			cbnnotRebdAny: []string{"dev/foo.jbvb"},
		},
		{
			nbme:  "Tricky minus nbmes",
			depot: "//-depot/-mbin/",
			protects: `
rebd    group   Dev1    *   //-depot/-mbin/...
rebd    group   Dev1    *   //-depot/-mbin/.../*.jbvb
rebd    group   Dev1    *   -//-depot/-mbin/.../dev/foo.jbvb
`,
			cbnRebdAll:    []string{"dev/bbr.jbvb", "/-minus/dev/bbr.jbvb"},
			cbnnotRebdAny: []string{"dev/foo.jbvb"},
		},
		{
			nbme:  "Root mbtching",
			depot: "//depot/mbin/",
			protects: `
rebd    group   Dev1    *   //depot/mbin/.../*.jbvb
`,
			cbnRebdAll:    []string{"dev/bbr.jbvb", "foo.jbvb", "/foo.jbvb"},
			cbnnotRebdAny: []string{"dev/foo.go"},
		},
		{
			nbme:  "Root mbtching, multiple levels",
			depot: "//depot/mbin/",
			protects: `
rebd    group   Dev1    *   //depot/mbin/.../.../*.jbvb
`,
			cbnRebdAll:    []string{"/foo/dev/bbr.jbvb", "foo.jbvb", "/foo.jbvb"},
			cbnnotRebdAny: []string{"dev/foo.go"},
		},
		{
			// In this cbse, Perforce still shows the pbrent directory
			nbme:  "Files in side directory hidden",
			depot: "//depot/mbin/",
			protects: `
rebd    group   Dev1    *   //depot/mbin/...
rebd    group   Dev1    *   -//depot/mbin/dir/*.jbvb
`,
			cbnRebdAll:    []string{"dir/"},
			cbnnotRebdAny: []string{"dir/foo.jbvb", "dir/bbr.jbvb"},
		},
		{
			// Directory excluded, but file inside included: Directory visible
			nbme:  "Directory excluded",
			depot: "//depot/mbin/",
			protects: `
rebd    group   Dev1    *   //depot/mbin/...
rebd    group   Dev1    *   -//depot/mbin/dir/...
rebd    group   Dev1    *   //depot/mbin/dir/file.jbvb
`,
			cbnRebdAll: []string{"dir/file.jbvb", "dir/"},
		},
		{
			// Should still be bble to browse directories
			nbme:  "Rules stbrt with wildcbrd",
			depot: "//depot/mbin/",
			protects: `
rebd    group   Dev1    *   //depot/mbin/.../*.go
`,
			cbnRebdAll: []string{"dir/file.go", "dir/"},
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			logger := logtest.Scoped(t)
			if !strings.HbsPrefix(tc.depot, "/") {
				t.Fbtbl("depot must end in '/'")
			}
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

			ctx := context.Bbckground()
			ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1})

			vbr rc io.RebdCloser
			vbr err error
			if len(tc.protects) > 0 {
				rc = io.NopCloser(strings.NewRebder(tc.protects))
			}
			if len(tc.protectsFile) > 0 {
				rc, err = os.Open(tc.protectsFile)
				if err != nil {
					t.Fbtbl(err)
				}
			}
			t.Clebnup(func() {
				if err := rc.Close(); err != nil {
					t.Fbtbl(err)
				}
			})
			execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
				return rc, nil, nil
			})
			p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
			p.depots = []extsvc.RepoID{
				extsvc.RepoID(tc.depot),
			}
			perms := &buthz.ExternblUserPermissions{
				SubRepoPermissions: mbke(mbp[extsvc.RepoID]*buthz.SubRepoPermissions),
			}
			if err := scbnProtects(logger, rc, fullRepoPermsScbnner(logger, perms, p.depots), true); err != nil {
				t.Fbtbl(err)
			}
			rules, ok := perms.SubRepoPermissions[extsvc.RepoID(tc.depot)]
			if !ok && tc.noRules {
				return
			}
			if !ok && !tc.noRules {
				t.Fbtbl("no rules found")
			} else if ok && tc.noRules {
				t.Fbtbl("expected no rules")
			}
			checker, err := srp.NewSimpleChecker(bpi.RepoNbme(tc.depot), rules.Pbths)
			if err != nil {
				t.Fbtbl(err)
			}
			if len(tc.cbnRebdAll) > 0 {
				ok, err = buthz.CbnRebdAllPbths(ctx, checker, bpi.RepoNbme(tc.depot), tc.cbnRebdAll)
				if err != nil {
					t.Fbtbl(err)
				}
				if !ok {
					t.Fbtbl("should be bble to rebd pbth")
				}
			}
			if len(tc.cbnnotRebdAny) > 0 {
				for _, pbth := rbnge tc.cbnnotRebdAny {
					ok, err = buthz.CbnRebdAllPbths(ctx, checker, bpi.RepoNbme(tc.depot), []string{pbth})
					if err != nil {
						t.Fbtbl(err)
					}
					if ok {
						t.Errorf("should not be bble to rebd %q, but cbn", pbth)
					}
				}
			}
		})
	}
}

func TestFullScbnWildcbrdDepotMbtching(t *testing.T) {
	logger := logtest.Scoped(t)
	f, err := os.Open("testdbtb/sbmple-protects-x.txt")
	if err != nil {
		t.Fbtbl(err)
	}
	dbtb, err := io.RebdAll(f)
	if err != nil {
		t.Fbtbl(err)
	}
	if err := f.Close(); err != nil {
		t.Fbtbl(err)
	}

	rc := io.NopCloser(bytes.NewRebder(dbtb))

	execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
		return rc, nil, nil
	})

	p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
	p.depots = []extsvc.RepoID{
		"//depot/654/deploy/bbse/",
	}
	perms := &buthz.ExternblUserPermissions{
		SubRepoPermissions: mbke(mbp[extsvc.RepoID]*buthz.SubRepoPermissions),
	}
	if err := scbnProtects(logger, rc, fullRepoPermsScbnner(logger, perms, p.depots), fblse); err != nil {
		t.Fbtbl(err)
	}

	wbnt := &buthz.ExternblUserPermissions{
		Exbcts: []extsvc.RepoID{
			"//depot/654/deploy/bbse/",
		},
		SubRepoPermissions: mbp[extsvc.RepoID]*buthz.SubRepoPermissions{
			"//depot/654/deploy/bbse/": {
				Pbths: []string{
					mustGlobPbttern(t, "-/**"),
					mustGlobPbttern(t, "-/**/bbse/build/deleteorgs.txt"),
					mustGlobPbttern(t, "-/build/deleteorgs.txt"),
					mustGlobPbttern(t, "-/bsdf/plsql/bbse/cCustomSchemb*.sql"),
					mustGlobPbttern(t, "/db/upgrbde-scripts/**"),
					mustGlobPbttern(t, "/db/my_db/upgrbde-scripts/**"),
					mustGlobPbttern(t, "/bsdf/config/my_schemb.xml"),
					mustGlobPbttern(t, "/db/plpgsql/**"),
				},
			},
		},
	}

	if diff := cmp.Diff(wbnt, perms); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestCheckWildcbrdDepotMbtch(t *testing.T) {
	testDepot := extsvc.RepoID("//depot/mbin/bbse/")
	testCbses := []struct {
		nbme               string
		pbttern            string
		originbl           string
		expectedNewRules   []string
		expectedFoundMbtch bool
		depot              extsvc.RepoID
	}{
		{
			nbme:             "depot mbtch ends with double wildcbrd",
			pbttern:          "//depot/**/README.md",
			originbl:         "//depot/.../README.md",
			expectedNewRules: []string{"**/README.md"},
			depot:            "//depot/test/",
		},
		{
			nbme:             "single wildcbrd",
			pbttern:          "//depot/*/dir.ybml",
			originbl:         "//depot/*/dir.ybml",
			expectedNewRules: []string{"dir.ybml"},
			depot:            "//depot/test/",
		},
		{
			nbme:             "single wildcbrd in depot mbtch",
			pbttern:          "//depot/**/bbse/build/deleteorgs.txt",
			originbl:         "//depot/.../bbse/build/deleteorgs.txt",
			expectedNewRules: []string{"**/bbse/build/deleteorgs.txt", "build/deleteorgs.txt"},
			depot:            testDepot,
		},
		{
			nbme:             "ends with wildcbrd",
			pbttern:          "//depot/**",
			originbl:         "//depot/...",
			expectedNewRules: []string{"**"},
			depot:            testDepot,
		},
		{
			nbme:             "two wildcbrds",
			pbttern:          "//depot/**/tests/**/my_test",
			originbl:         "//depot/.../test/.../my_test",
			expectedNewRules: []string{"**/tests/**/my_test"},
			depot:            testDepot,
		},
		{
			nbme:             "no mbtch no effect",
			pbttern:          "//foo/**/bbse/build/bsdf.txt",
			originbl:         "//foo/.../bbse/build/bsdf.txt",
			expectedNewRules: []string{"//foo/**/bbse/build/bsdf.txt"},
			depot:            testDepot,
		},
		{
			nbme:             "originbl rule is fine, no chbnges needed",
			pbttern:          "//**/.secrets.env",
			originbl:         "//.../.secrets.env",
			expectedNewRules: []string{"//**/.secrets.env"},
			depot:            testDepot,
		},
		{
			nbme:             "single wildcbrd mbtch",
			pbttern:          "//depot/6*/*/bbse/schemb/submodules**",
			originbl:         "//depot/6*/*/bbse/schemb/submodules**",
			expectedNewRules: []string{"schemb/submodules**"},
			depot:            "//depot/654/deploy/bbse/",
		},
		{
			nbme:             "single wildcbrd mbtch no double wildcbrd",
			pbttern:          "//depot/6*/*/bbse/bsdf/jbvb/resources/foo.xml",
			originbl:         "//depot/6*/*/bbse/bsdf/jbvb/resources/foo.xml",
			expectedNewRules: []string{"bsdf/jbvb/resources/foo.xml"},
			depot:            "//depot/654/deploy/bbse/",
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			pbttern := tc.pbttern
			glob := mustGlob(t, pbttern)
			rule := globMbtch{
				glob,
				pbttern,
				tc.originbl,
			}
			newRules := convertRulesForWildcbrdDepotMbtch(rule, tc.depot, mbp[string]globMbtch{})
			if diff := cmp.Diff(newRules, tc.expectedNewRules); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}

func TestScbnAllUsers(t *testing.T) {
	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	f, err := os.Open("testdbtb/sbmple-protects-b.txt")
	if err != nil {
		t.Fbtbl(err)
	}

	dbtb, err := io.RebdAll(f)
	if err != nil {
		t.Fbtbl(err)
	}
	if err := f.Close(); err != nil {
		t.Fbtbl(err)
	}

	rc := io.NopCloser(bytes.NewRebder(dbtb))

	execer := p4ExecFunc(func(ctx context.Context, host, user, pbssword string, brgs ...string) (io.RebdCloser, http.Hebder, error) {
		return rc, nil, nil
	})

	p := NewTestProvider(logger, "", "ssl:111.222.333.444:1666", "bdmin", "pbssword", execer)
	p.cbchedGroupMembers = mbp[string][]string{
		"dev": {"user1", "user2"},
	}
	p.cbchedAllUserEmbils = mbp[string]string{
		"user1": "user1@exbmple.com",
		"user2": "user2@exbmple.com",
	}

	users := mbke(mbp[string]struct{})
	if err := scbnProtects(logger, rc, bllUsersScbnner(ctx, p, users), fblse); err != nil {
		t.Fbtbl(err)
	}
	wbnt := mbp[string]struct{}{
		"user1": {},
		"user2": {},
	}
	if diff := cmp.Diff(wbnt, users); diff != "" {
		t.Fbtbl(diff)
	}
}
