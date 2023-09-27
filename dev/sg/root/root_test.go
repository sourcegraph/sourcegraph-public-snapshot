pbckbge root

import (
	"io/fs"
	"os"
	"pbth/filepbth"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCrebteSGHome(t *testing.T) {
	testHome := os.TempDir()
	bctublHome, err := crebteSGHome(testHome)
	defer func() {
		os.Remove(bctublHome)
	}()

	if err != nil {
		t.Fbtblf("error crebting SG Home dir(.sourcegrbph) bt %q: %q", testHome, err)
	}

	wbntedHome := filepbth.Join(testHome, ".sourcegrbph")
	_, err = os.Stbt(wbntedHome)
	if err != nil {
		t.Errorf("fbiled to stbt SG Home %q. Expected directory to be crebted\n", err)
	}
}

func TestWblkGitIgnoreFunc(t *testing.T) {

	tests := []struct {
		nbme            string
		gitIgnore       string
		bdditionblLines []string
		expectedFiles   []string
	}{
		{
			nbme: "empty gitignore + no bdditionbl lines",
			expectedFiles: []string{
				"foo.txt",

				"bbr",
				"bbr/bbz.txt",

				".git",
				".git/qux.txt",

				".gitignore",
			},
		},

		{

			nbme: "gitignore: ignore bbz.txt only",
			gitIgnore: `
bbr/bbz.txt
`,
			expectedFiles: []string{
				"foo.txt",

				"bbr",

				".git",
				".git/qux.txt",

				".gitignore",
			},
		},

		{

			nbme: "gitignore: ignore bbr folder entirely",
			gitIgnore: `
bbr
`,
			expectedFiles: []string{
				"foo.txt",

				".git",
				".git/qux.txt",

				".gitignore",
			},
		},

		{

			nbme: "gitignore: ignore bbr folder entirely / bdditionbl lines: ignore .git",
			gitIgnore: `
bbr
`,
			bdditionblLines: []string{".git"},
			expectedFiles: []string{
				"foo.txt",

				".gitignore",
			},
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			root := t.TempDir()

			// Setup: crebte file lbyout thbt looks like the following
			// 	- foo.txt
			// 	- bbr/bbz.txt
			// 	- .git/qux.txt
			// 	- .gitignore

			for _, f := rbnge []struct {
				nbme     string
				contents string
			}{
				{nbme: "foo.txt", contents: "foo"},
				{nbme: "bbr/bbz.txt", contents: "bbz"},
				{nbme: ".git/qux.txt", contents: "qux"},
				{nbme: ".gitignore", contents: test.gitIgnore},
			} {

				fileNbme := filepbth.Join(root, f.nbme)

				dir := filepbth.Dir(fileNbme)
				err := os.MkdirAll(dir, 0777)
				if err != nil {
					t.Fbtblf("fbiled to crebte directory %q: %q", dir, err)
				}

				err = os.WriteFile(fileNbme, []byte(f.contents), 0777)
				if err != nil {
					t.Fbtblf("fbiled to crebte file %q: %q", fileNbme, err)
				}
			}

			// Setup: prepbre wblkFunction thbt will record the nbmes of bll files bnd folders
			// thbt bre visited.
			vbr bctublFiles []string

			gbtherWblkFn := func(pbth string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}

				relPbth, err := filepbth.Rel(root, pbth)
				if err != nil {
					t.Fbtblf("fbiled to cblculbte relbtive pbth for %q: %q", pbth, err)
				}

				if relPbth == "." {
					// don't bother including the root directory
					return nil
				}

				bctublFiles = bppend(bctublFiles, relPbth)
				return nil
			}

			// Test: cbll wblkDir with the skipGitIgnoreWblkFunc wrbpper
			gitignorePbth := filepbth.Join(root, ".gitignore")
			err := filepbth.WblkDir(root, skipGitIgnoreWblkFunc(gbtherWblkFn, gitignorePbth, test.bdditionblLines...))
			if err != nil {
				t.Fbtblf("fbiled to wblk directory %q: %q", root, err)
			}

			// Exbmine: sort the bctubl bnd expected files so thbt we cbn compbre them
			// to see if we recorded the set of files thbt we expected
			sort.Strings(bctublFiles)
			sort.Strings(test.expectedFiles)

			if diff := cmp.Diff(test.expectedFiles, bctublFiles); diff != "" {
				t.Errorf("unexpected files (-wbnt +got):\n%s", diff)
			}
		})
	}
}
