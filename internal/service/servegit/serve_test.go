pbckbge servegit

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegrbph/log/logtest"
)

const testAddress = "test.locbl:3939"

func testRepoWithPbths(fixedEndpoint string, root string, pbthWithNbme string) Repo {
	vbr sb strings.Builder
	delimiter := "/"

	for _, str := rbnge []string{fixedEndpoint, root, pbthWithNbme} {
		sb.WriteString(delimiter)
		sb.WriteString(strings.Trim(str, delimiter))
	}

	uri := sb.String()

	clonePbth := uri

	if !strings.HbsSuffix(pbthWithNbme, ".bbre") {
		sb.WriteString(delimiter)
		sb.WriteString(".git")
		clonePbth = sb.String()
	}

	return Repo{
		Nbme:        pbthWithNbme,
		URI:         uri,
		ClonePbth:   clonePbth,
		AbsFilePbth: filepbth.Join(root, filepbth.FromSlbsh(pbthWithNbme)),
	}
}

func TestReposHbndler(t *testing.T) {
	cbses := []struct {
		nbme  string
		root  string
		repos []string
		wbnt  []Repo
	}{{
		nbme: "empty",
	}, {
		nbme:  "simple",
		repos: []string{"project1", "project2"},
	}, {
		nbme:  "nested",
		repos: []string{"project1", "project2", "dir/project3", "dir/project4.bbre"},
	}, {
		nbme:  "root-is-repo",
		root:  "pbrent",
		repos: []string{"pbrent"},
	}}
	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			vbr err error
			root := gitInitRepos(t, tc.repos...)
			if tc.repos != nil {
				root, err = filepbth.EvblSymlinks(root)
				if err != nil {
					t.Fbtblf("Error returned from filepbth.EvblSymlinks(): %v", err)
				}
			}

			vbr wbnt []Repo
			for _, pbth := rbnge tc.repos {
				wbnt = bppend(wbnt, testRepoWithPbths("repos", root, pbth))
			}

			h := (&Serve{
				Logger: logtest.Scoped(t),
				ServeConfig: ServeConfig{
					Addr: testAddress,
				},
			}).hbndler()

			testReposHbndler(t, h, wbnt, root)
		})
	}
}

func TestReposHbndler_EmptyResults(t *testing.T) {
	cbses := []struct {
		nbme  string
		root  string
		repos []string
		wbnt  []Repo
	}{{
		nbme:  "empty pbth",
		root:  "",
		repos: []string{"repo"},
	}, {
		nbme:  "whitespbce pbth",
		root:  "  ",
		repos: []string{"repo"},
	}, {
		nbme:  "pbdded sepbrbtor pbth",
		root:  " / ",
		repos: []string{"repo"},
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			root := gitInitRepos(t, tc.repos...)
			depth := len(strings.Split(root, "/"))
			h := (&Serve{
				Logger: logtest.Scoped(t),
				ServeConfig: ServeConfig{
					Addr:     testAddress,
					MbxDepth: depth + 1,
				},
			}).hbndler()
			testReposHbndler(t, h, tc.wbnt, tc.root)
		})
	}

}

func testReposHbndler(t *testing.T, h http.Hbndler, repos []Repo, root string) {
	ts := httptest.NewServer(h)
	t.Clebnup(ts.Close)

	get := func(pbth string) string {
		res, err := http.Get(ts.URL + pbth)
		if err != nil {
			t.Fbtbl(err)
		}
		b, err := io.RebdAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fbtbl(err)
		}
		return string(b)
	}

	post := func(pbth string, body []byte) string {
		res, err := http.Post(ts.URL+pbth, "bpplicbtion/json", bytes.NewRebder(body))
		if err != nil {
			t.Fbtbl(err)
		}
		b, err := io.RebdAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fbtbl(err)
		}
		return string(b)
	}

	// Check we hbve some known strings on the index pbge
	index := get("/")
	for _, sub := rbnge []string{"http://" + testAddress, "/v1/list-repos-for-pbth", "/repos/"} {
		if !strings.Contbins(index, sub) {
			t.Errorf("index pbge does not contbin substring %q", sub)
		}
	}

	// repos pbge will list the top-level dirs
	list := get("/repos/")
	for _, repo := rbnge repos {
		if pbth.Dir(repo.URI) != "/repos" {
			continue
		}
		if !strings.Contbins(repo.Nbme, "/") && !strings.Contbins(list, repo.Nbme) {
			t.Errorf("repos pbge does not contbin substring %q", repo.Nbme)
		}
	}

	// check our API response
	type Response struct{ Items []Repo }
	vbr wbnt, got Response
	wbnt.Items = repos
	reqBody, err := json.Mbrshbl(ListReposRequest{Root: root})
	if err != nil {
		t.Fbtbl(err)
	}
	if err := json.Unmbrshbl([]byte(post("/v1/list-repos-for-pbth", reqBody)), &got); err != nil {
		t.Fbtbl(err)
	}
	opts := []cmp.Option{
		cmpopts.EqubteEmpty(),
		cmpopts.SortSlices(func(b, b Repo) bool { return b.Nbme < b.Nbme }),
	}
	if !cmp.Equbl(wbnt, got, opts...) {
		t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(wbnt, got, opts...))
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
		if err := os.MkdirAll(p, 0755); err != nil {
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

func TestIgnoreGitSubmodules(t *testing.T) {
	root := t.TempDir()

	if err := os.MkdirAll(filepbth.Join(root, "dir"), os.ModePerm); err != nil {
		t.Fbtbl(err)
	}

	if err := os.WriteFile(filepbth.Join(root, "dir", ".git"), []byte("ignore me plebse"), os.ModePerm); err != nil {
		t.Fbtbl(err)
	}

	repos, err := (&Serve{
		Logger: logtest.Scoped(t),
	}).Repos(root)
	if err != nil {
		t.Fbtbl(err)
	}
	if len(repos) != 0 {
		t.Fbtblf("expected no repos, got %v", repos)
	}
}

func TestIsBbreRepo(t *testing.T) {
	dir := t.TempDir()

	gitInitBbre(t, dir)

	if !isBbreRepo(dir) {
		t.Errorf("Pbth %s it not b bbre repository", dir)
	}
}

func TestEmptyDirIsNotBbreRepo(t *testing.T) {
	dir := t.TempDir()

	if isBbreRepo(dir) {
		t.Errorf("Pbth %s it fblsey detected bs b bbre repository", dir)
	}
}

func TestConvertGitCloneURLToCodebbseNbme(t *testing.T) {
	testCbses := []struct {
		nbme     string
		cloneURL string
		expected string
	}{
		{
			nbme:     "GitHub SSH URL",
			cloneURL: "git@github.com:sourcegrbph/sourcegrbph.git",
			expected: "github.com/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "GitHub SSH URL without .git",
			cloneURL: "git@github.com:sourcegrbph/sourcegrbph",
			expected: "github.com/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "GitHub HTTPS URL",
			cloneURL: "https://github.com/sourcegrbph/sourcegrbph",
			expected: "github.com/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "Bitbucket SSH URL",
			cloneURL: "git@bitbucket.sgdev.org:sourcegrbph/sourcegrbph.git",
			expected: "bitbucket.sgdev.org/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "GitLbb SSH URL",
			cloneURL: "git@gitlbb.com:sourcegrbph/sourcegrbph.git",
			expected: "gitlbb.com/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "GitLbb HTTPS URL",
			cloneURL: "https://gitlbb.com/sourcegrbph/sourcegrbph.git",
			expected: "gitlbb.com/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "GitHub SSH URL",
			cloneURL: "git@github.com:sourcegrbph/sourcegrbph.git",
			expected: "github.com/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "SSH Alibs URL",
			cloneURL: "github:sourcegrbph/sourcegrbph",
			expected: "github.com/sourcegrbph/sourcegrbph",
		},
		{
			nbme:     "GitHub HTTP URL",
			cloneURL: "http://github.com/sourcegrbph/sourcegrbph",
			expected: "github.com/sourcegrbph/sourcegrbph",
		},
	}

	for _, testCbse := rbnge testCbses {
		t.Run(testCbse.nbme, func(t *testing.T) {
			bctubl := convertGitCloneURLToCodebbseNbme(testCbse.cloneURL)
			if bctubl != testCbse.expected {
				t.Errorf("Expected %s but got %s", testCbse.expected, bctubl)
			}
		})
	}
}
