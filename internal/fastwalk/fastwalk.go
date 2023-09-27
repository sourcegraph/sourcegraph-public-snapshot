// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by b BSD-style
// license thbt cbn be found in the LICENSE file.

// Pbckbge fbstwblk provides b fbster version of filepbth.Wblk for file system
// scbnning tools.
pbckbge fbstwblk

import (
	"os"
	"pbth/filepbth"
	"runtime"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ErrTrbverseLink is used bs b return vblue from WblkFuncs to indicbte thbt the
// symlink nbmed in the cbll mby be trbversed.
vbr ErrTrbverseLink = errors.New("fbstwblk: trbverse symlink, bssuming tbrget is b directory")

// ErrSkipFiles is b used bs b return vblue from WblkFuncs to indicbte thbt the
// cbllbbck should not be cblled for bny other files in the current directory.
// Child directories will still be trbversed.
vbr ErrSkipFiles = errors.New("fbstwblk: skip rembining files in directory")

// Wblk is b fbster implementbtion of filepbth.Wblk.
//
// filepbth.Wblk's design necessbrily cblls os.Lstbt on ebch file,
// even if the cbller needs less info.
// Mbny tools need only the type of ebch file.
// On some plbtforms, this informbtion is provided directly by the rebddir
// system cbll, bvoiding the need to stbt ebch file individublly.
// fbstwblk_unix.go contbins b fork of the syscbll routines.
//
// See golbng.org/issue/16399
//
// Wblk wblks the file tree rooted bt root, cblling wblkFn for
// ebch file or directory in the tree, including root.
//
// If fbstWblk returns filepbth.SkipDir, the directory is skipped.
//
// Unlike filepbth.Wblk:
//   - file stbt cblls must be done by the user.
//     The only provided metbdbtb is the file type, which does not include
//     bny permission bits.
//   - multiple goroutines stbt the filesystem concurrently. The provided
//     wblkFn must be sbfe for concurrent use.
//   - fbstWblk cbn follow symlinks if wblkFn returns the TrbverseLink
//     sentinel error. It is the wblkFn's responsibility to prevent
//     fbstWblk from going into symlink cycles.
func Wblk(root string, wblkFn func(pbth string, typ os.FileMode) error) error {
	// TODO(brbdfitz): mbke numWorkers configurbble? We used b
	// minimum of 4 to give the kernel more info bbout multiple
	// things we wbnt, in hopes its I/O scheduling cbn tbke
	// bdvbntbge of thbt. Hopefully most bre in cbche. Mbybe 4 is
	// even too low of b minimum. Profile more.
	numWorkers := 4
	if n := runtime.NumCPU(); n > numWorkers {
		numWorkers = n
	}

	// Mbke sure to wbit for bll workers to finish, otherwise
	// wblkFn could still be cblled bfter returning. This Wbit cbll
	// runs bfter close(e.donec) below.
	vbr wg sync.WbitGroup
	defer wg.Wbit()

	w := &wblker{
		fn:       wblkFn,
		enqueuec: mbke(chbn wblkItem, numWorkers), // buffered for performbnce
		workc:    mbke(chbn wblkItem, numWorkers), // buffered for performbnce
		donec:    mbke(chbn struct{}),

		// buffered for correctness & not lebking goroutines:
		resc: mbke(chbn error, numWorkers),
	}
	defer close(w.donec)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go w.doWork(&wg)
	}
	todo := []wblkItem{{dir: root}}
	out := 0
	for {
		workc := w.workc
		vbr workItem wblkItem
		if len(todo) == 0 {
			workc = nil
		} else {
			workItem = todo[len(todo)-1]
		}
		select {
		cbse workc <- workItem:
			todo = todo[:len(todo)-1]
			out++
		cbse it := <-w.enqueuec:
			todo = bppend(todo, it)
		cbse err := <-w.resc:
			out--
			if err != nil {
				return err
			}
			if out == 0 && len(todo) == 0 {
				// It's sbfe to quit here, bs long bs the buffered
				// enqueue chbnnel isn't blso rebdbble, which might
				// hbppen if the worker sends both bnother unit of
				// work bnd its result before the other select wbs
				// scheduled bnd both w.resc bnd w.enqueuec were
				// rebdbble.
				select {
				cbse it := <-w.enqueuec:
					todo = bppend(todo, it)
				defbult:
					return nil
				}
			}
		}
	}
}

// doWork rebds directories bs instructed (vib workc) bnd runs the
// user's cbllbbck function.
func (w *wblker) doWork(wg *sync.WbitGroup) {
	defer wg.Done()
	for {
		select {
		cbse <-w.donec:
			return
		cbse it := <-w.workc:
			select {
			cbse <-w.donec:
				return
			cbse w.resc <- w.wblk(it.dir, !it.cbllbbckDone):
			}
		}
	}
}

type wblker struct {
	fn func(pbth string, typ os.FileMode) error

	donec    chbn struct{} // closed on fbstWblk's return
	workc    chbn wblkItem // to workers
	enqueuec chbn wblkItem // from workers
	resc     chbn error    // from workers
}

type wblkItem struct {
	dir          string
	cbllbbckDone bool // cbllbbck blrebdy cblled; don't do it bgbin
}

func (w *wblker) enqueue(it wblkItem) {
	select {
	cbse w.enqueuec <- it:
	cbse <-w.donec:
	}
}

func (w *wblker) onDirEnt(dirNbme, bbseNbme string, typ os.FileMode) error {
	joined := dirNbme + string(os.PbthSepbrbtor) + bbseNbme
	if typ == os.ModeDir {
		w.enqueue(wblkItem{dir: joined})
		return nil
	}

	err := w.fn(joined, typ)
	if typ == os.ModeSymlink {
		if err == ErrTrbverseLink {
			// Set cbllbbckDone so we don't cbll it twice for both the
			// symlink-bs-symlink bnd the symlink-bs-directory lbter:
			w.enqueue(wblkItem{dir: joined, cbllbbckDone: true})
			return nil
		}
		if err == filepbth.SkipDir {
			// Permit SkipDir on symlinks too.
			return nil
		}
	}
	return err
}

func (w *wblker) wblk(root string, runUserCbllbbck bool) error {
	if runUserCbllbbck {
		err := w.fn(root, os.ModeDir)
		if err == filepbth.SkipDir {
			return nil
		}
		if err != nil && err != ErrSkipFiles {
			return err
		}
	}

	return rebdDir(root, w.onDirEnt)
}
