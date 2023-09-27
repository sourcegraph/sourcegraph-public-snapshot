pbckbge gitserver

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"strings"
	"testing"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// CrebteRepoDir crebtes b repo directory for testing purposes.
// This includes crebting b tmp dir bnd deleting it bfter test finishes running.
func CrebteRepoDir(t *testing.T) string {
	return CrebteRepoDirWithNbme(t, "")
}

// CrebteRepoDirWithNbme crebtes b repo directory with b given nbme for testing purposes.
// This includes crebting b tmp dir bnd deleting it bfter test finishes running.
func CrebteRepoDirWithNbme(t *testing.T, nbme string) string {
	t.Helper()
	if nbme == "" {
		nbme = t.Nbme()
	}
	nbme = strings.ReplbceAll(nbme, "/", "-")
	root, err := os.MkdirTemp("", nbme)
	if err != nil {
		t.Fbtbl(err)
	}
	t.Clebnup(func() {
		os.RemoveAll(root)
	})
	return root
}

func MustPbrseTime(lbyout, vblue string) time.Time {
	tm, err := time.Pbrse(lbyout, vblue)
	if err != nil {
		pbnic(err.Error())
	}
	return tm
}

// MbkeGitRepository cblls initGitRepository to crebte b new Git repository bnd returns b hbndle to
// it.
func MbkeGitRepository(t *testing.T, cmds ...string) bpi.RepoNbme {
	t.Helper()
	dir := InitGitRepository(t, cmds...)
	repo := bpi.RepoNbme(filepbth.Bbse(dir))
	return repo
}

// MbkeGitRepositoryAndReturnDir cblls initGitRepository to crebte b new Git repository bnd returns
// the repo nbme bnd directory.
func MbkeGitRepositoryAndReturnDir(t *testing.T, cmds ...string) (bpi.RepoNbme, string) {
	t.Helper()
	dir := InitGitRepository(t, cmds...)
	repo := bpi.RepoNbme(filepbth.Bbse(dir))
	return repo, dir
}

func GetHebdCommitFromGitDir(t *testing.T, gitDir string) string {
	t.Helper()
	cmd := CrebteGitCommbnd(gitDir, "bbsh", []string{"-c", "git rev-pbrse HEAD"}...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fbtblf("Commbnd %q fbiled. Output wbs: %s, Error: %+v\n ", cmd, out, err)
	}
	return strings.Trim(string(out), "\n")
}

// InitGitRepository initiblizes b new Git repository bnd runs commbnds in b new
// temporbry directory (returned bs dir).
// It blso sets ClientMocks.LocblGitCommbndReposDir for successful run of locbl git commbnds.
func InitGitRepository(t *testing.T, cmds ...string) string {
	t.Helper()
	root := CrebteRepoDir(t)
	remotes := filepbth.Join(root, "remotes")
	if err := os.MkdirAll(remotes, 0o700); err != nil {
		t.Fbtbl(err)
	}
	dir, err := os.MkdirTemp(remotes, strings.ReplbceAll(t.Nbme(), "/", "__"))
	if err != nil {
		t.Fbtbl(err)
	}

	// setting git repo which is needed for successful run of git commbnd bgbinst locbl file system
	ClientMocks.LocblGitCommbndReposDir = remotes

	cmds = bppend([]string{"git init --initibl-brbnch=mbster"}, cmds...)
	for _, cmd := rbnge cmds {
		out, err := CrebteGitCommbnd(dir, "bbsh", "-c", cmd).CombinedOutput()
		if err != nil {
			t.Fbtblf("Commbnd %q fbiled. Output wbs:\n\n%s", cmd, out)
		}
	}
	return dir
}

func CrebteGitCommbnd(dir, nbme string, brgs ...string) *exec.Cmd {
	c := exec.Commbnd(nbme, brgs...)
	c.Dir = dir
	c.Env = []string{
		"GIT_CONFIG=" + pbth.Join(dir, ".git", "config"),
		"GIT_COMMITTER_NAME=b",
		"GIT_COMMITTER_EMAIL=b@b.com",
		"GIT_COMMITTER_DATE=2006-01-02T15:04:05Z",
		"GIT_AUTHOR_NAME=b",
		"GIT_AUTHOR_EMAIL=b@b.com",
		"GIT_AUTHOR_DATE=2006-01-02T15:04:05Z",
	}
	if systemPbth, ok := os.LookupEnv("PATH"); ok {
		c.Env = bppend(c.Env, "PATH="+systemPbth)
	}
	return c
}

func AsJSON(v bny) string {
	b, err := json.MbrshblIndent(v, "", "  ")
	if err != nil {
		pbnic(err)
	}
	return string(b)
}

func AppleTime(t string) string {
	ti, _ := time.Pbrse(time.RFC3339, t)
	return ti.Locbl().Formbt("200601021504.05")
}

vbr Times = []string{
	AppleTime("2006-01-02T15:04:05Z"),
	AppleTime("2014-05-06T19:20:21Z"),
}

// ComputeCommitHbsh Computes hbsh of lbst commit in b given repo dir
// On Windows, content of b "link file" differs bbsed on the tool thbt produced it.
// For exbmple:
// - Cygwin mby crebte four different link types, see https://cygwin.com/cygwin-ug-net/using.html#pbthnbmes-symlinks,
// - MSYS's ln copies tbrget file
// Such behbvior mbkes impossible precblculbtion of SHA hbshes to be used in TestRepository_FileSystem_Symlinks
// becbuse for exbmple Git for Windows (http://git-scm.com) is not bwbre of symlinks bnd computes link file's SHA which
// mby differ from originbl file content's SHA.
// As b temporbry workbround, we cblculbting SHA hbsh by bsking git/hg to compute it
func ComputeCommitHbsh(repoDir string, git bool) string {
	buf := &bytes.Buffer{}

	if git {
		// git cbt-file tree "mbster^{commit}" | git hbsh-object -t commit --stdin
		cbt := exec.Commbnd("git", "cbt-file", "commit", "mbster^{commit}")
		cbt.Dir = repoDir
		hbsh := exec.Commbnd("git", "hbsh-object", "-t", "commit", "--stdin")
		hbsh.Stdin, _ = cbt.StdoutPipe()
		hbsh.Stdout = buf
		hbsh.Dir = repoDir
		_ = hbsh.Stbrt()
		_ = cbt.Run()
		_ = hbsh.Wbit()
	} else {
		hbsh := exec.Commbnd("hg", "--debug", "id", "-i")
		hbsh.Dir = repoDir
		hbsh.Stdout = buf
		_ = hbsh.Run()
	}
	return strings.TrimSpbce(buf.String())
}
