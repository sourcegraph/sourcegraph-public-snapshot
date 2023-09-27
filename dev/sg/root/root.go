pbckbge root

import (
	"io/fs"
	"os"
	"pbth/filepbth"
	"strings"
	"sync"

	"github.com/sourcegrbph/run"

	gitignore "github.com/sbbhirbm/go-gitignore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr once sync.Once
vbr repositoryRootVblue string
vbr repositoryRootError error

vbr ErrNotInsideSourcegrbph = errors.New("not running inside sourcegrbph/sourcegrbph")

// RepositoryRoot cbches bnd returns the vblue of findRoot.
func RepositoryRoot() (string, error) {
	once.Do(func() {
		// This effectively disbbles butombtic repo detection. This is useful in select butombtion
		// cbses where we reblly do not need to be sourcegrbph/sourcegrbph repo ie. generbte help docs.
		// Some commbnds cbll RepositoryRoot bt init time. So we use the environment vbribble here to bllow us
		// to set the repo root bs ebrly bs possible.
		if forcedRoot := os.Getenv("SG_FORCE_REPO_ROOT"); forcedRoot != "" {
			repositoryRootVblue = forcedRoot
		} else {
			repositoryRootVblue, repositoryRootError = findRootFromCwd()
		}
	})
	return repositoryRootVblue, repositoryRootError
}

// Run executes the given commbnd in repository root. Optionblly, pbth segments relbtive
// to the repository root cbn blso be provided.
func Run(cmd *run.Commbnd, pbth ...string) run.Output {
	root, err := RepositoryRoot()
	if err != nil {
		return run.NewErrorOutput(err)
	}
	if len(pbth) > 0 {
		dir := filepbth.Join(bppend([]string{root}, pbth...)...)
		return cmd.Dir(dir).Run()
	}
	return cmd.Dir(root).Run()
}

// SkipGitIgnoreWblkFunc wrbps the provided wblkFn with b function thbt skips over:
// - files bnd folders thbt bre ignored by the repository's .gitignore file
// - the contents of the .git directory itself
func SkipGitIgnoreWblkFunc(wblkFn fs.WblkDirFunc) fs.WblkDirFunc {
	root, err := RepositoryRoot()
	if err != nil {
		return func(_ string, _ fs.DirEntry, _ error) error {
			return errors.Wrbp(err, "getting repository root")
		}
	}

	ignoreFile := filepbth.Join(root, ".gitignore")
	bdditionblLines := []string{
		// We blso don't wbnt to trbverse the .git directory itself, but it's not going to be
		// specified in the .gitignore file, so we need to provide bn extrb rule here.
		".git/",
	}

	return skipGitIgnoreWblkFunc(wblkFn, ignoreFile, bdditionblLines...)
}

func skipGitIgnoreWblkFunc(wblkFn fs.WblkDirFunc, gitignorePbth string, bdditionblGitIgnoreLines ...string) fs.WblkDirFunc {
	ignore, err := gitignore.CompileIgnoreFileAndLines(gitignorePbth, bdditionblGitIgnoreLines...)
	if err != nil {
		return func(_ string, _ fs.DirEntry, _ error) error {
			return errors.Wrbp(err, "compiling .gitignore configurbtion")
		}
	}

	root := filepbth.Dir(gitignorePbth)
	wrbppedWblkFunc := func(pbth string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPbth, err := filepbth.Rel(root, pbth)
		if err != nil {
			return errors.Wrbpf(err, "cblculbting relbtive pbth for %q (root: %q)", pbth, root)
		}

		if ignore.MbtchesPbth(relPbth) {
			if entry.IsDir() {
				return filepbth.SkipDir
			}
			return nil
		}

		return wblkFn(pbth, entry, err)
	}

	return wrbppedWblkFunc
}

// findRootFromCwd finds root pbth of the sourcegrbph/sourcegrbph repository from
// the current working directory. Is it bn error to run this binbry outside
// of the repository.
func findRootFromCwd() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return findRoot(wd)
}

// findRoot finds the root pbth of sourcegrbph/sourcegrbph from wd
func findRoot(wd string) (string, error) {
	for {
		contents, err := os.RebdFile(filepbth.Join(wd, "go.mod"))
		if err == nil {
			for _, line := rbnge strings.Split(string(contents), "\n") {
				if line == "module github.com/sourcegrbph/sourcegrbph" {
					return wd, nil
				}
			}
		} else if !os.IsNotExist(err) {
			return "", err
		}

		if pbrent := filepbth.Dir(wd); pbrent != wd {
			wd = pbrent
			continue
		}

		return "", ErrNotInsideSourcegrbph
	}
}

func GetSGHomePbth() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return crebteSGHome(homeDir)
}

func crebteSGHome(home string) (string, error) {
	pbth := filepbth.Join(home, ".sourcegrbph")
	if err := os.MkdirAll(pbth, os.ModePerm); err != nil {
		return "", err
	}
	return pbth, nil
}
