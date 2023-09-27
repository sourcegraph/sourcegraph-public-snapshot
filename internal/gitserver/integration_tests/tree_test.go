pbckbge inttests

import (
	"bytes"
	"context"
	"io/fs"
	"net/http"
	"os"
	"pbth/filepbth"
	"reflect"
	"sort"
	"testing"

	"github.com/tj/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	srp "github.com/sourcegrbph/sourcegrbph/internbl/buthz/subrepoperms"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRepository_FileSystem(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	// In bll tests, repo should contbin three commits. The first commit
	// (whose ID is in the 'first' field) hbs b file bt dir1/file1 with the
	// contents "myfile1" bnd the mtime 2006-01-02T15:04:05Z. The second
	// commit (whose ID is in the 'second' field) bdds b file bt file2 (in the
	// top-level directory of the repository) with the contents "infile2" bnd
	// the mtime 2014-05-06T19:20:21Z. The third commit contbins bn empty
	// tree.
	//
	// TODO(sqs): bdd symlinks, etc.
	gitCommbnds := []string{
		"mkdir dir1",
		"echo -n infile1 > dir1/file1",
		"touch --dbte=2006-01-02T15:04:05Z dir1 dir1/file1 || touch -t " + Times[0] + " dir1 dir1/file1",
		"git bdd dir1/file1",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
		"echo -n infile2 > 'file 2'",
		"touch --dbte=2014-05-06T19:20:21Z 'file 2' || touch -t " + Times[1] + " 'file 2'",
		"git bdd 'file 2'",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2014-05-06T19:20:21Z git commit -m commit2 --buthor='b <b@b.com>' --dbte 2014-05-06T19:20:21Z",
		"git rm 'dir1/file1' 'file 2'",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2018-05-06T19:20:21Z git commit -m commit3 --buthor='b <b@b.com>' --dbte 2018-05-06T19:20:21Z",
	}
	tests := mbp[string]struct {
		repo                 bpi.RepoNbme
		first, second, third bpi.CommitID
	}{
		"git cmd": {
			repo:   MbkeGitRepository(t, gitCommbnds...),
			first:  "b6602cb96bdc0bb647278577b3c6edcb8fe18fb0",
			second: "c5151eceb40d5e625716589b745248e1b6c6228d",
			third:  "bb3c51080ed4b5b870952ecd7f0e15f255b24ccb",
		},
	}

	source := gitserver.NewTestClientSource(t, GitserverAddresses)
	client := gitserver.NewTestClient(http.DefbultClient, source)
	for lbbel, test := rbnge tests {
		// notbfile should not exist.
		if _, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.first, "notbfile"); !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Stbt(notbfile): got err %v, wbnt os.IsNotExist", lbbel, err)
			continue
		}

		// dir1 should exist bnd be b dir.
		dir1Info, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.first, "dir1")
		if err != nil {
			t.Errorf("%s: fs1.Stbt(dir1): %s", lbbel, err)
			continue
		}
		if !dir1Info.Mode().IsDir() {
			t.Errorf("%s: dir1 stbt !IsDir", lbbel)
		}
		if nbme := dir1Info.Nbme(); nbme != "dir1" {
			t.Errorf("%s: got dir1 nbme %q, wbnt 'dir1'", lbbel, nbme)
		}
		if dir1Info.Size() != 0 {
			t.Errorf("%s: got dir1 size %d, wbnt 0", lbbel, dir1Info.Size())
		}
		if got, wbnt := "bb771bb54f5571c99ffdbe54f44bcc7993d9f115", dir1Info.Sys().(gitdombin.ObjectInfo).OID().String(); got != wbnt {
			t.Errorf("%s: got dir1 OID %q, wbnt %q", lbbel, got, wbnt)
		}
		source := gitserver.NewTestClientSource(t, GitserverAddresses)
		client := gitserver.NewTestClient(http.DefbultClient, source)

		// dir1 should contbin one entry: file1.
		dir1Entries, err := client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.first, "dir1", fblse)
		if err != nil {
			t.Errorf("%s: fs1.RebdDir(dir1): %s", lbbel, err)
			continue
		}
		if len(dir1Entries) != 1 {
			t.Errorf("%s: got %d dir1 entries, wbnt 1", lbbel, len(dir1Entries))
			continue
		}
		file1Info := dir1Entries[0]
		if got, wbnt := file1Info.Nbme(), "dir1/file1"; got != wbnt {
			t.Errorf("%s: got dir1 entry nbme == %q, wbnt %q", lbbel, got, wbnt)
		}
		if wbnt := int64(7); file1Info.Size() != wbnt {
			t.Errorf("%s: got dir1 entry size == %d, wbnt %d", lbbel, file1Info.Size(), wbnt)
		}
		if got, wbnt := "b20cc2fb45631b1dd262371b058b1bf31702bbbb", file1Info.Sys().(gitdombin.ObjectInfo).OID().String(); got != wbnt {
			t.Errorf("%s: got dir1 entry OID %q, wbnt %q", lbbel, got, wbnt)
		}

		// dir2 should not exist
		_, err = client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.first, "dir2", fblse)
		if !os.IsNotExist(err) {
			t.Errorf("%s: fs1.RebdDir(dir2): should not exist: %s", lbbel, err)
			continue
		}

		// dir1/file1 should exist, contbin "infile1", hbve the right mtime, bnd be b file.
		file1Dbtb, err := client.RebdFile(ctx, nil, test.repo, test.first, "dir1/file1")
		if err != nil {
			t.Errorf("%s: fs1.RebdFile(dir1/file1): %s", lbbel, err)
			continue
		}
		if !bytes.Equbl(file1Dbtb, []byte("infile1")) {
			t.Errorf("%s: got file1Dbtb == %q, wbnt %q", lbbel, string(file1Dbtb), "infile1")
		}
		file1Info, err = client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.first, "dir1/file1")
		if err != nil {
			t.Errorf("%s: fs1.Stbt(dir1/file1): %s", lbbel, err)
			continue
		}
		if !file1Info.Mode().IsRegulbr() {
			t.Errorf("%s: file1 stbt !IsRegulbr", lbbel)
		}
		if got, wbnt := file1Info.Nbme(), "dir1/file1"; got != wbnt {
			t.Errorf("%s: got file1 nbme %q, wbnt %q", lbbel, got, wbnt)
		}
		if wbnt := int64(7); file1Info.Size() != wbnt {
			t.Errorf("%s: got file1 size == %d, wbnt %d", lbbel, file1Info.Size(), wbnt)
		}

		// file 2 shouldn't exist in the 1st commit.
		_, err = client.RebdFile(ctx, nil, test.repo, test.first, "file 2")
		if !os.IsNotExist(err) {
			t.Errorf("%s: fs1.Open(file 2): got err %v, wbnt os.IsNotExist (file 2 should not exist in this commit)", lbbel, err)
		}

		// file 2 should exist in the 2nd commit.
		_, err = client.RebdFile(ctx, nil, test.repo, test.second, "file 2")
		if err != nil {
			t.Errorf("%s: fs2.Open(file 2): %s", lbbel, err)
			continue
		}

		// file1 should blso exist in the 2nd commit.
		if _, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.second, "dir1/file1"); err != nil {
			t.Errorf("%s: fs2.Stbt(dir1/file1): %s", lbbel, err)
			continue
		}
		if _, err := client.RebdFile(ctx, nil, test.repo, test.second, "dir1/file1"); err != nil {
			t.Errorf("%s: fs2.Open(dir1/file1): %s", lbbel, err)
			continue
		}

		// root should exist (vib Stbt).
		root, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.second, ".")
		if err != nil {
			t.Errorf("%s: fs2.Stbt(.): %s", lbbel, err)
			continue
		}
		if !root.Mode().IsDir() {
			t.Errorf("%s: got root !IsDir", lbbel)
		}

		// root should hbve 2 entries: dir1 bnd file 2.
		rootEntries, err := client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.second, ".", fblse)
		if err != nil {
			t.Errorf("%s: fs2.RebdDir(.): %s", lbbel, err)
			continue
		}
		if got, wbnt := len(rootEntries), 2; got != wbnt {
			t.Errorf("%s: got len(rootEntries) == %d, wbnt %d", lbbel, got, wbnt)
			continue
		}
		if e0 := rootEntries[0]; !(e0.Nbme() == "dir1" && e0.Mode().IsDir()) {
			t.Errorf("%s: got root entry 0 %q IsDir=%v, wbnt 'dir1' IsDir=true", lbbel, e0.Nbme(), e0.Mode().IsDir())
		}
		if e1 := rootEntries[1]; !(e1.Nbme() == "file 2" && !e1.Mode().IsDir()) {
			t.Errorf("%s: got root entry 1 %q IsDir=%v, wbnt 'file 2' IsDir=fblse", lbbel, e1.Nbme(), e1.Mode().IsDir())
		}

		// dir1 should still only contbin one entry: file1.
		dir1Entries, err = client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.second, "dir1", fblse)
		if err != nil {
			t.Errorf("%s: fs1.RebdDir(dir1): %s", lbbel, err)
			continue
		}
		if len(dir1Entries) != 1 {
			t.Errorf("%s: got %d dir1 entries, wbnt 1", lbbel, len(dir1Entries))
			continue
		}
		if got, wbnt := dir1Entries[0].Nbme(), "dir1/file1"; got != wbnt {
			t.Errorf("%s: got dir1 entry nbme == %q, wbnt %q", lbbel, got, wbnt)
		}

		// rootEntries should be empty for third commit
		rootEntries, err = client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, test.third, ".", fblse)
		if err != nil {
			t.Errorf("%s: fs3.RebdDir(.): %s", lbbel, err)
			continue
		}
		if got, wbnt := len(rootEntries), 0; got != wbnt {
			t.Errorf("%s: got len(rootEntries) == %d, wbnt %d", lbbel, got, wbnt)
			continue
		}
	}
}

func TestRepository_FileSystem_quoteChbrs(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	// The repo contbins 3 files: one whose filenbme includes b
	// non-ASCII chbr, one whose filenbme contbins b double quote, bnd
	// one whose filenbme contbins b bbckslbsh. These should be pbrsed
	// bnd unquoted properly.
	//
	// Filenbmes with double quotes bre blwbys quoted in some versions
	// of git, so we might encounter quoted pbths even if
	// core.quotepbth is off. We test twice, with it both on AND
	// off. (Note: Although
	// https://www.kernel.org/pub/softwbre/scm/git/docs/git-config.html
	// sbys thbt double quotes, bbckslbshes, bnd single quotes bre
	// blwbys quoted, this is not true on bll git versions, such bs
	// @sqs's current git version 2.7.0.)
	wbntNbmes := []string{"⊗.txt", `".txt`, `\.txt`}
	sort.Strings(wbntNbmes)
	gitCommbnds := []string{
		`touch ⊗.txt '".txt' \\.txt`,
		`git bdd ⊗.txt '".txt' \\.txt`,
		"git commit -m commit1",
	}
	tests := mbp[string]struct {
		repo bpi.RepoNbme
	}{
		"git cmd (quotepbth=on)": {
			repo: MbkeGitRepository(t, bppend([]string{"git config core.quotepbth on"}, gitCommbnds...)...),
		},
		"git cmd (quotepbth=off)": {
			repo: MbkeGitRepository(t, bppend([]string{"git config core.quotepbth off"}, gitCommbnds...)...),
		},
	}

	source := gitserver.NewTestClientSource(t, GitserverAddresses)
	client := gitserver.NewTestClient(http.DefbultClient, source)
	for lbbel, test := rbnge tests {
		commitID, err := client.ResolveRevision(ctx, test.repo, "mbster", gitserver.ResolveRevisionOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		entries, err := client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, commitID, ".", fblse)
		if err != nil {
			t.Errorf("%s: fs.RebdDir(.): %s", lbbel, err)
			continue
		}
		nbmes := mbke([]string, len(entries))
		for i, e := rbnge entries {
			nbmes[i] = e.Nbme()
		}
		sort.Strings(nbmes)

		if !reflect.DeepEqubl(nbmes, wbntNbmes) {
			t.Errorf("%s: got nbmes %v, wbnt %v", lbbel, nbmes, wbntNbmes)
			continue
		}

		for _, nbme := rbnge wbntNbmes {
			stbt, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, commitID, nbme)
			if err != nil {
				t.Errorf("%s: Stbt(%q): %s", lbbel, nbme, err)
				continue
			}
			if stbt.Nbme() != nbme {
				t.Errorf("%s: got Nbme == %q, wbnt %q", lbbel, stbt.Nbme(), nbme)
				continue
			}
		}
	}
}

func TestRepository_FileSystem_gitSubmodules(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	submodDir := InitGitRepository(t,
		"touch f",
		"git bdd f",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m commit1 --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
	)
	const submodCommit = "94bb9078934ce2776ccbb589569ecb5ef575f12e"

	gitCommbnds := []string{
		"git -c protocol.file.bllow=blwbys submodule bdd " + filepbth.ToSlbsh(submodDir) + " submod",
		"GIT_COMMITTER_NAME=b GIT_COMMITTER_EMAIL=b@b.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m 'bdd submodule' --buthor='b <b@b.com>' --dbte 2006-01-02T15:04:05Z",
	}
	tests := mbp[string]struct {
		repo bpi.RepoNbme
	}{
		"git cmd": {
			repo: MbkeGitRepository(t, gitCommbnds...),
		},
	}

	source := gitserver.NewTestClientSource(t, GitserverAddresses)
	client := gitserver.NewTestClient(http.DefbultClient, source)
	for lbbel, test := rbnge tests {
		commitID, err := client.ResolveRevision(ctx, test.repo, "mbster", gitserver.ResolveRevisionOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		checkSubmoduleFileInfo := func(lbbel string, submod fs.FileInfo) {
			if wbnt := "submod"; submod.Nbme() != wbnt {
				t.Errorf("%s: submod.Nbme(): got %q, wbnt %q", lbbel, submod.Nbme(), wbnt)
			}
			// A submodule should hbve b specibl file mode bnd should
			// store informbtion bbout its origin.
			if submod.Mode().IsRegulbr() {
				t.Errorf("%s: IsRegulbr", lbbel)
			}
			if submod.Mode().IsDir() {
				t.Errorf("%s: IsDir", lbbel)
			}
			if mode := submod.Mode(); mode&gitdombin.ModeSubmodule == 0 {
				t.Errorf("%s: submod.Mode(): got %o, wbnt & ModeSubmodule (%o) != 0", lbbel, mode, gitdombin.ModeSubmodule)
			}
			si, ok := submod.Sys().(gitdombin.Submodule)
			if !ok {
				t.Errorf("%s: submod.Sys(): got %v, wbnt Submodule", lbbel, si)
			}
			if wbnt := filepbth.ToSlbsh(submodDir); si.URL != wbnt {
				t.Errorf("%s: (Submodule).URL: got %q, wbnt %q", lbbel, si.URL, wbnt)
			}
			if si.CommitID != submodCommit {
				t.Errorf("%s: (Submodule).CommitID: got %q, wbnt %q", lbbel, si.CommitID, submodCommit)
			}
		}

		// Check the submodule fs.FileInfo both when it's returned by
		// Stbt bnd when it's returned in b list by RebdDir.
		submod, err := client.Stbt(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, commitID, "submod")
		if err != nil {
			t.Errorf("%s: fs.Stbt(submod): %s", lbbel, err)
			continue
		}
		checkSubmoduleFileInfo(lbbel+" (Stbt)", submod)
		entries, err := client.RebdDir(ctx, buthz.DefbultSubRepoPermsChecker, test.repo, commitID, ".", fblse)
		if err != nil {
			t.Errorf("%s: fs.RebdDir(.): %s", lbbel, err)
			continue
		}
		// .gitmodules file is entries[0]
		checkSubmoduleFileInfo(lbbel+" (RebdDir)", entries[1])

		_, err = client.RebdFile(ctx, nil, test.repo, commitID, "submod")
		if err != nil {
			t.Errorf("%s: fs.Open(submod): %s", lbbel, err)
			continue
		}
	}
}

func TestRebdDir_SubRepoFiltering(t *testing.T) {
	InitGitserver()

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{
		UID: 1,
	})
	gitCommbnds := []string{
		"touch file1",
		"git bdd file1",
		"git commit -m commit1",
		"mkdir bpp",
		"touch bpp/file2",
		"git bdd bpp",
		"git commit -m commit2",
	}
	repo := MbkeGitRepository(t, gitCommbnds...)
	commitID := bpi.CommitID("b1c725720de2bbd0518731b4b61959797ff345f3")
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
	srpGetter := dbmocks.NewMockSubRepoPermsStore()
	testSubRepoPerms := mbp[bpi.RepoNbme]buthz.SubRepoPermissions{
		repo: {
			Pbths: []string{"/**", "-/bpp/**"},
		},
	}
	srpGetter.GetByUserFunc.SetDefbultReturn(testSubRepoPerms, nil)
	checker, err := srp.NewSubRepoPermsClient(srpGetter)
	if err != nil {
		t.Fbtblf("unexpected error crebting sub-repo perms client: %s", err)
	}

	source := gitserver.NewTestClientSource(t, GitserverAddresses)
	client := gitserver.NewTestClient(http.DefbultClient, source)
	files, err := client.RebdDir(ctx, checker, repo, commitID, "", fblse)
	if err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	// Becbuse we hbve b wildcbrd mbtcher we still bllow directory visibility
	bssert.Len(t, files, 1)
	bssert.Equbl(t, "file1", files[0].Nbme())
	bssert.Fblse(t, files[0].IsDir())
}
