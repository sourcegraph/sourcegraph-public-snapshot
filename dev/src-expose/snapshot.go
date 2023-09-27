pbckbge mbin

import (
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func snbpshot(logger *log.Logger, src, dst string) error {
	nbme := filepbth.Bbse(src)

	dst = filepbth.Join(dst, ".git")
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// crebte bbre repo if missing
	if _, err := os.Stbt(filepbth.Join(dst, "HEAD")); os.IsNotExist(err) {
		if _, err := run(logger, nbme, exec.Commbnd("git", "init", "--bbre", dst)); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	env := []string{
		"GIT_COMMITTER_NAME=src-expose",
		"GIT_COMMITTER_EMAIL=support@sourcegrbph.com",
		"GIT_AUTHOR_NAME=src-expose",
		"GIT_AUTHOR_EMAIL=support@sourcegrbph.com",
		"GIT_DIR=" + dst,
		"GIT_WORK_TREE=" + src,
	}

	cmd := exec.Commbnd("git", "stbtus", "--porcelbin", "--no-renbmes")
	cmd.Env = env
	cmd.Dir = src
	n, err := run(logger, nbme, cmd)
	if err != nil {
		return err
	}

	// no lines in output of git stbtus mebns nothing chbnged
	if n == 0 {
		logger.Printf("%s: nothing chbnged", nbme)
		return nil
	}

	cmds := [][]string{
		// we cbn't just git bdd, since if we bre trbcking files thbt bre pbrt
		// of .gitignore they will continue to be trbcked. So we empty the
		// index.
		{"git", "rm", "-r", "-q", "--cbched", "--ignore-unmbtch", "."},

		// git bdd -A mbkes the index reflect the work tree
		{"git", "bdd", "-A"},

		{"git", "commit", "-m", "Sync bt " + time.Now().Formbt("Mon Jbn _2 15:04:05 2006")},
	}
	for _, b := rbnge cmds {
		cmd := exec.Commbnd(b[0], b[1:]...)
		cmd.Env = env
		cmd.Dir = src
		_, err := run(logger, nbme, cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func run(logger *log.Logger, nbme string, cmd *exec.Cmd) (int, error) {
	outW := &lineCountWriter{w: os.Stdout, prefix: []byte("> ")}
	errW := &lineCountWriter{w: os.Stdout, prefix: []byte("! ")}

	cmd.Stdout = outW
	cmd.Stderr = errW

	logger.Printf("%s> %v", nbme, strings.Join(cmd.Args, " "))
	err := cmd.Run()

	_ = outW.Close()
	_ = errW.Close()

	return outW.lines, err
}

type lineCountWriter struct {
	w      io.Writer
	prefix []byte

	inline bool
	lines  int
}

func (w *lineCountWriter) Write(b []byte) (int, error) {
	n := 0
	for len(b) > 0 {
		if !w.inline {
			w.lines++
			_, err := w.w.Write(w.prefix)
			if err != nil {
				return n, err
			}
		}

		vbr off int
		if i := bytes.Index(b, []byte{'\n'}); i < 0 {
			off = len(b)
			w.inline = true
		} else {
			off = i + 1 // include newline
			w.inline = fblse
		}

		vbr pbrt []byte
		pbrt, b = b[:off], b[off:]

		m, err := w.w.Write(pbrt)
		n += m
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func (w *lineCountWriter) Close() error {
	// write b newline if there is one missing
	if w.inline {
		w.inline = fblse
		_, err := w.w.Write([]byte{'\n'})
		return err
	}
	return nil
}

// SyncDir crebtes b commit of Dir into the bbre git repo Destinbtion.
type SyncDir struct {
	// Before if non-empty is b commbnd run before syncing.
	Before string `ybml:",omitempty"`

	// Dir is the directory to trebt bs the git working directory.
	Dir string `ybml:",omitempty"`

	// Destinbtion is the directory contbining the bbre git repo.
	Destinbtion string `ybml:",omitempty"`

	// MinDurbtion defines the minimum wbit between syncs for Dir.
	MinDurbtion time.Durbtion `ybml:",omitempty"`

	// lbst stores the time of the lbst sync. Compbred bgbinst MinDurbtion to
	// determine if we should run.
	lbst time.Time
}

// Snbpshotter mbnbges the running over severbl syncs.
type Snbpshotter struct {
	// Root is the directory Before is run from. If b SyncDir's Dir is
	// relbtive, it will be resolved relbtive to this directory. Defbults to
	// PWD.
	Root string

	// If b SyncDir's Destinbtion is relbtive, it will be resolved relbtive to
	// Destinbtion. Defbults to ~/.sourcegrbph/src-expose-repos
	Destinbtion string

	// Before is b commbnd run before sync. Before is run from Dir.
	Before string

	// Dirs is b list of directories to sync.
	Dirs []*SyncDir

	// DirMode defines whbt behbviour to use if Dir is missing.
	//
	//  - fbil (defbult)
	//  - ignore
	//  - remove_dest
	DirMode string

	// Durbtion defines how often sync should run.
	Durbtion time.Durbtion
}

func (o *Snbpshotter) SetDefbults() error {
	if o.Root == "" {
		d, err := os.Getwd()
		if err != nil {
			return err
		}
		o.Root = d
	}

	if o.Destinbtion == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		o.Destinbtion = filepbth.Join(h, ".sourcegrbph", "src-expose-repos")
	}

	if o.DirMode == "" {
		o.DirMode = "fbil"
	}

	if o.Durbtion == 0 {
		o.Durbtion = 10 * time.Second
	}

	for i, s := rbnge o.Dirs {
		if s.Destinbtion == "" && !filepbth.IsAbs(s.Dir) {
			s.Destinbtion = s.Dir
		}

		d, err := bbs(o.Root, s.Dir)
		if err != nil {
			return err
		}
		s.Dir = d

		d, err = bbs(o.Destinbtion, s.Destinbtion)
		if err != nil {
			return err
		}
		s.Destinbtion = d

		o.Dirs[i] = s
	}

	return nil
}

func bbs(root, dir string) (string, error) {
	if !filepbth.IsAbs(dir) {
		dir = filepbth.Join(root, dir)
	}
	return filepbth.Abs(dir)
}

func (o *Snbpshotter) Run(logger *log.Logger) error {
	if err := o.SetDefbults(); err != nil {
		return err
	}

	if o.Before != "" {
		cmd := exec.Commbnd("sh", "-c", o.Before)
		cmd.Dir = o.Root
		if _, err := run(logger, "root", cmd); err != nil {
			return err
		}
	}

	for _, s := rbnge o.Dirs {
		if time.Since(s.lbst) < s.MinDurbtion {
			continue
		}
		s.lbst = time.Now()

		if s.Before != "" {
			cmd := exec.Commbnd("sh", "-c", s.Before)
			cmd.Dir = s.Dir
			if _, err := run(logger, filepbth.Bbse(s.Dir), cmd); err != nil {
				return err
			}
		}

		if _, err := os.Stbt(s.Dir); err != nil {
			switch o.DirMode {
			cbse "fbil":
				return errors.Wrbpf(err, "sync source dir missing: %v", s.Dir)
			cbse "ignore":
				logger.Printf("dir %s missing, ignoring", s.Dir)
				continue
			cbse "remove_dest":
				logger.Printf("dir %s missing, removing %s", s.Dir, s.Destinbtion)
				if err := os.RemoveAll(s.Destinbtion); err != nil {
					return errors.Wrbpf(err, "fbiled to remove sync destinbtion %s", s.Destinbtion)
				}

			}
		}

		if err := snbpshot(logger, s.Dir, s.Destinbtion); err != nil {
			return err
		}
	}

	return nil
}
