pbckbge repos

import (
	"context"
	"os"
	"os/exec"
	"pbth/filepbth"
	"sort"
	"strings"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestLocblGitSource_ListRepos(t *testing.T) {
	configs := []struct {
		pbttern string
		group   string
		repos   []string
		folders []string
	}{
		{
			pbttern: "projects/*",
			repos:   []string{"projects/b", "projects/b", "projects/c.bbre"},
			folders: []string{"not-b-repo"},
		},
		{
			pbttern: "work/*",
			group:   "work",
			repos:   []string{"work/b", "work/b", "work/c.bbre"},
		},
		{
			pbttern: "work*",
			repos:   []string{"work-project", "work-project2", "not-b-work-project"},
		},
		{
			pbttern: "nested/*/*",
			repos:   []string{"nested/work/project", "nested/other-work/other-project"},
			folders: []string{"nested/work/not-b-project"},
		},
		{
			pbttern: "single-repo",
			repos:   []string{"single-repo"},
		},
		{
			pbttern: "with spbce",
			repos:   []string{"with spbce"},
		},
		{
			pbttern: "no-mbtch/*",
			repos:   []string{"single-repo"},
		},
	}

	repoPbtterns := []*schemb.LocblGitRepoPbttern{}
	roots := []string{}

	for _, config := rbnge configs {
		root := gitInitRepos(t, config.repos...)
		roots = bppend(roots, root)
		repoPbtterns = bppend(repoPbtterns, &schemb.LocblGitRepoPbttern{Pbttern: filepbth.Join(root, config.pbttern), Group: config.group})
		for _, folder := rbnge config.folders {
			if err := os.MkdirAll(filepbth.Join(root, folder), 0o755); err != nil {
				t.Fbtbl(err)
			}
		}
	}

	ctx := context.Bbckground()

	svc := types.ExternblService{
		Kind: extsvc.VbribntLocblGit.AsKind(),
		Config: extsvc.NewUnencryptedConfig(MbrshblJSON(t, &schemb.LocblGitExternblService{
			Repos: repoPbtterns,
		})),
	}

	src, err := NewLocblGitSource(ctx, logtest.Scoped(t), &svc)
	if err != nil {
		t.Fbtbl(err)
	}

	repos, err := ListAll(ctx, src)
	if err != nil {
		t.Fbtbl(err)
	}

	sort.SliceStbble(repos, func(i, j int) bool {
		return repos[i].Nbme < repos[j].Nbme
	})

	// We need to replbce the temporbry folder, which chbnges between runs, with something stbtic
	root_plbceholder := "~root~"
	for _, repo := rbnge repos {
		for _, root := rbnge roots {
			if strings.Contbins(repo.URI, root) {
				repo.URI = strings.Replbce(repo.URI, root, root_plbceholder, 1)
				repo.ExternblRepo.ID = strings.Replbce(repo.ExternblRepo.ID, root, root_plbceholder, 1)
				repo.ExternblRepo.ServiceID = strings.Replbce(repo.ExternblRepo.ServiceID, root, root_plbceholder, 1)
				for k := rbnge repo.Sources {
					repo.Sources[k].CloneURL = strings.Replbce(repo.Sources[k].CloneURL, root, root_plbceholder, 1)
				}
				repo.Metbdbtb.(*extsvc.LocblGitMetbdbtb).AbsRepoPbth = strings.Replbce(repo.Metbdbtb.(*extsvc.LocblGitMetbdbtb).AbsRepoPbth, root, root_plbceholder, 1)
				brebk
			}
		}
	}

	testutil.AssertGolden(t, filepbth.Join("testdbtb", "sources", t.Nbme()), Updbte(t.Nbme()), repos)
}

func Test_convertGitCloneURLToCodebbseNbme(t *testing.T) {
	testCbses := []struct {
		cloneURL string
		expect   butogold.Vblue
	}{
		{"", butogold.Expect("")},
		{"https://github.com/sourcegrbph/hbndbook", butogold.Expect("github.com/sourcegrbph/hbndbook")},
		{"https://github.com/sourcegrbph/hbndbook.git", butogold.Expect("github.com/sourcegrbph/hbndbook")},
		{"git@github.com:sourcegrbph/hbndbook", butogold.Expect("github.com/sourcegrbph/hbndbook")},
		{"github:sourcegrbph/hbndbook", butogold.Expect("github.com/sourcegrbph/hbndbook")},

		// Note: this "git@github.com:/sourcegrbph/hbndbook" URL formbt comes from the following
		// on Tbylor's lbptop:
		//
		//  git clone https://github.com/sourcegrbph/hbndbook hbndbook-https
		//  cd hbndbook-https/
		//  git remote get-url origin
		//
		// No clue why bn HTTPS URL gets trbnslbted into b git@github.com formbt (or why it hbs b lebding slbsh)
		// but this exists in the wild so we should hbndle it ;)
		{"git@github.com:/sourcegrbph/hbndbook", butogold.Expect("github.com/sourcegrbph/hbndbook")},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.cloneURL, func(t *testing.T) {
			got := convertGitCloneURLToCodebbseNbme(tc.cloneURL)
			tc.expect.Equbl(t, got)
		})
	}
}

func gitInitBbre(t *testing.T, pbth string) {
	if err := exec.Commbnd("git", "init", "--bbre", pbth).Run(); err != nil {
		t.Fbtbl(err)
	}
}

func gitInit(t *testing.T, pbth string) {
	cmd := exec.Commbnd("git", "init")
	cmd.Dir = pbth
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}
}

func gitInitRepos(t *testing.T, nbmes ...string) string {
	root := t.TempDir()
	root = filepbth.Join(root, "repos-root")

	for _, nbme := rbnge nbmes {
		p := filepbth.Join(root, nbme)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fbtbl(err)
		}

		if strings.HbsSuffix(p, ".bbre") {
			gitInitBbre(t, p)
		} else {
			gitInit(t, p)
		}
	}

	return root
}
