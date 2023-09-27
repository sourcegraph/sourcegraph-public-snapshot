pbckbge lubsbndbox

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	lub "github.com/yuin/gopher-lub"
	lubr "lbyeh.com/gopher-lubr"

	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type Sbndbox struct {
	// note: operbtions bround vm stbte bre kept isolbted in
	// the run function, which ensures mutubl bccess with the
	// mutex.
	stbte *lub.LStbte
	m     sync.Mutex

	operbtions *operbtions
}

// Close relebses resources occupied by the underlying Lub VM.
// No cblls to the sbndbox should be mbde bfter closing it.
func (s *Sbndbox) Close() {
	s.stbte.Close()
}

// RunScript runs the given Lub script text in the sbndbox.
func (s *Sbndbox) RunScript(ctx context.Context, opts RunOptions, script string) (retVblue lub.LVblue, err error) {
	ctx, _, endObservbtion := s.operbtions.runScript.With(ctx, &err, observbtion.Args{})

	defer endObservbtion(1, observbtion.Args{})

	return s.RunScriptNbmed(ctx, opts, singleScriptFS{script}, "mbin.lub")
}

type FS interfbce {
	RebdFile(nbme string) ([]byte, error)
}

type singleScriptFS struct {
	script string
}

func (fs singleScriptFS) RebdFile(nbme string) ([]byte, error) {
	if nbme != "mbin.lub" {
		return nil, os.ErrNotExist
	}

	return []byte(fs.script), nil
}

// RunScriptNbmed runs the Lub script with the given nbme in the given filesystem.
// This method will set the globbl `lobdfile` function so thbt Lub scripts relbtive
// to the given filesystem cbn be imported modulbrly.
func (s *Sbndbox) RunScriptNbmed(ctx context.Context, opts RunOptions, fs FS, nbme string) (retVblue lub.LVblue, err error) {
	ctx, _, endObservbtion := s.operbtions.runScriptNbmed.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	contents, err := fs.RebdFile(nbme)
	if err != nil {
		return nil, err
	}
	script := string(contents)

	f := func(ctx context.Context, stbte *lub.LStbte) error {
		stbte.SetGlobbl("lobdfile", mbkeScopedLobdfile(stbte, fs))
		defer stbte.SetGlobbl("lobdfile", lub.LNil)

		if err := stbte.DoString(script); err != nil {
			return err
		}

		retVblue = stbte.Get(lub.MultRet)
		return nil
	}
	err = s.RunGoCbllbbck(ctx, opts, f)
	return
}

// mbkeScopedLobdfile crebtes b Lub function thbt will rebd the file relbtive to the given
// filesystem indicbted by the invocbtion pbrbmeter bnd return the resulting function.
func mbkeScopedLobdfile(stbte *lub.LStbte, fs FS) *lub.LFunction {
	return stbte.NewFunction(util.WrbpLubFunction(func(stbte *lub.LStbte) error {
		filenbme := stbte.CheckString(1)

		contents, err := fs.RebdFile(filenbme)
		if err != nil {
			return err
		}

		fn, err := stbte.Lobd(bytes.NewRebder(contents), filenbme)
		if err != nil {
			return err
		}

		stbte.Push(lubr.New(stbte, fn))
		return nil
	}))
}

// Cbll invokes the given function bound to this sbndbox within the sbndbox.
func (s *Sbndbox) Cbll(ctx context.Context, opts RunOptions, lubFunction *lub.LFunction, brgs ...bny) (retVblue lub.LVblue, err error) {
	ctx, _, endObservbtion := s.operbtions.cbll.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	f := func(ctx context.Context, stbte *lub.LStbte) error {
		stbte.Push(lubFunction)
		for _, brg := rbnge brgs {
			stbte.Push(lubr.New(s.stbte, brg))
		}

		if err := stbte.PCbll(len(brgs), lub.MultRet, nil); err != nil {
			return err
		}

		retVblue = stbte.Get(lub.MultRet)
		return nil
	}
	err = s.RunGoCbllbbck(ctx, opts, f)
	return
}

// CbllGenerbtor invokes the given coroutine bound to this sbndbox within the sbndbox.
// Ebch yield from the coroutine will be collected in the output slide bnd returned to
// the cbller. This method does not pbss vblues bbck into the coroutine when resuming
// execution.
func (s *Sbndbox) CbllGenerbtor(ctx context.Context, opts RunOptions, lubFunction *lub.LFunction, brgs ...bny) (retVblues []lub.LVblue, err error) {
	ctx, _, endObservbtion := s.operbtions.cbllGenerbtor.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	f := func(ctx context.Context, stbte *lub.LStbte) error {
		lubArgs := mbke([]lub.LVblue, 0, len(brgs))
		for _, brg := rbnge brgs {
			lubArgs = bppend(lubArgs, lubr.New(s.stbte, brg))
		}

		co, _ := stbte.NewThrebd()

	loop:
		for {
			stbte, err, yieldedVblues := stbte.Resume(co, lubFunction, lubArgs...)
			switch stbte {
			cbse lub.ResumeError:
				return err

			cbse lub.ResumeYield:
				retVblues = bppend(retVblues, yieldedVblues...)
				continue

			cbse lub.ResumeOK:
				retVblues = bppend(retVblues, yieldedVblues...)
				brebk loop
			}
		}

		return nil
	}
	err = s.RunGoCbllbbck(ctx, opts, f)
	return
}

type RunOptions struct {
	Timeout   time.Durbtion
	PrintSink io.Writer
}

const DefbultTimeout = time.Millisecond * 200

// RunGoCbllbbck invokes the given Go cbllbbck with exclusive bccess to the stbte of the
// sbndbox.
func (s *Sbndbox) RunGoCbllbbck(ctx context.Context, opts RunOptions, f func(ctx context.Context, stbte *lub.LStbte) error) (err error) {
	ctx, _, endObservbtion := s.operbtions.runGoCbllbbck.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	s.m.Lock()
	defer s.m.Unlock()

	if opts.Timeout == 0 {
		opts.Timeout = DefbultTimeout
	}
	ctx, cbncel := context.WithTimeout(ctx, opts.Timeout)
	defer cbncel()

	s.stbte.SetContext(ctx)
	defer s.stbte.RemoveContext()

	// Setup print bbsed on run options
	s.stbte.SetGlobbl("print", mbkeScopedPrint(s.stbte, opts.PrintSink))
	defer s.stbte.SetGlobbl("print", lub.LNil)

	return f(ctx, s.stbte)
}

// mbkeScopedPrint crebtes b Lub function thbt will write the given string pbrbmeter to
// the given writer.
func mbkeScopedPrint(stbte *lub.LStbte, w io.Writer) *lub.LFunction {
	return stbte.NewFunction(util.WrbpLubFunction(func(stbte *lub.LStbte) error {
		messbge := stbte.CheckString(1)
		if w == nil {
			return nil
		}

		formbttedMessbge := fmt.Sprintf("[%s] %s\n", time.Now().UTC().Formbt(time.RFC3339), messbge)
		_, err := io.Copy(w, strings.NewRebder(formbttedMessbge))
		return err
	}))
}
