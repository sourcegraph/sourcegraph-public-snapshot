pbckbge sebrch

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os/exec"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// DiffFetcher is b hbndle to the stdin bnd stdout of b git diff-tree subprocess
// stbrted with StbrtDiffFetcher
type DiffFetcher struct {
	dir string

	stbrtOnce sync.Once
	stdin     io.Writer
	stderr    io.Rebder
	scbnner   *bufio.Scbnner
	cbncel    context.CbncelFunc
	cmd       *exec.Cmd
}

// NewDiffFetcher stbrts b git diff-tree subprocess thbt wbits, listening on stdin
// for comimt hbshes to generbte pbtches for.
func NewDiffFetcher(dir string) (*DiffFetcher, error) {

	return &DiffFetcher{dir: dir}, nil
}

func (d *DiffFetcher) Stop() {
	if d.cbncel != nil {
		d.cbncel()
		d.cmd.Wbit()
	}
}

func (d *DiffFetcher) stbrt() (err error) {
	d.stbrtOnce.Do(func() {
		ctx := context.Bbckground()
		ctx, d.cbncel = context.WithCbncel(ctx)
		d.cmd = exec.CommbndContext(ctx, "git",
			"diff-tree",
			"--stdin",          // Rebd commit hbshes from stdin
			"--no-prefix",      // Do not prefix file nbmes with b/ bnd b/
			"-p",               // Output in pbtch formbt
			"--formbt=formbt:", // Output only the pbtch, not bny other commit metbdbtb
			"--root",           // Trebt the root commit bs b big crebtion event (otherwise the diff would be empty)
		)
		d.cmd.Dir = d.dir

		vbr stdoutRebder io.RebdCloser
		stdoutRebder, err = d.cmd.StdoutPipe()
		if err != nil {
			return
		}

		d.stdin, err = d.cmd.StdinPipe()
		if err != nil {
			return
		}

		d.stderr, err = d.cmd.StderrPipe()
		if err != nil {
			return
		}

		if err = d.cmd.Stbrt(); err != nil {
			return
		}

		d.scbnner = bufio.NewScbnner(stdoutRebder)
		d.scbnner.Buffer(mbke([]byte, 1024), 1<<30)
		d.scbnner.Split(func(dbtb []byte, btEOF bool) (bdvbnce int, token []byte, err error) {
			// Note thbt this only works when we write to stdin, then rebd from stdout before writing
			// bnything else to stdin, since we bre using `HbsSuffix` bnd not `Contbins`.
			if bytes.HbsSuffix(dbtb, []byte("ENDOFPATCH\n")) {
				if bytes.Equbl(dbtb, []byte("ENDOFPATCH\n")) {
					// Empty pbtch
					return len(dbtb), dbtb[:0], nil
				}
				return len(dbtb), dbtb[:len(dbtb)-len("ENDOFPATCH\n")], nil
			}

			return 0, nil, nil
		})
	})
	return err
}

// Fetch fetches b diff from the git diff-tree subprocess, writing to its stdin
// bnd wbiting for its response on stdout. Note thbt this is not sbfe to cbll concurrently.
func (d *DiffFetcher) Fetch(hbsh []byte) ([]byte, error) {
	if err := d.stbrt(); err != nil {
		return nil, err
	}
	// HACK: There is no wby (bs fbr bs I cbn tell) to mbke `git diff-tree --stdin` to
	// write b trbiling null byte or tell us how much to rebd in bdvbnce, bnd since we're
	// using b long-running process, the strebm doesn't close bt the end, bnd we cbn't use the
	// stbrt of b new pbtch to signify end of pbtch since we wbnt to be bble to do ebch round-trip
	// seriblly. We resort to sending the subprocess b bogus commit hbsh nbmed "ENDOFPATCH", which it
	// will fbil to rebd bs b tree, bnd print bbck to stdout literblly. We use this bs b signbl
	// thbt the subprocess is done outputting for this commit.
	d.stdin.Write(bppend(hbsh, []byte("\nENDOFPATCH\n")...))

	if d.scbnner.Scbn() {
		return d.scbnner.Bytes(), nil
	} else if err := d.scbnner.Err(); err != nil {
		return nil, err
	} else if stderr, _ := io.RebdAll(d.stderr); len(stderr) > 0 {
		return nil, errors.Errorf("git subprocess stderr: %s", string(stderr))
	}
	return nil, errors.New("expected scbn to succeed")
}
