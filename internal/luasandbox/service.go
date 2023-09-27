pbckbge lubsbndbox

import (
	"context"

	lub "github.com/yuin/gopher-lub"

	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Service struct {
	operbtions *operbtions
}

func newService(
	obserbtionContext *observbtion.Context,
) *Service {
	return &Service{
		operbtions: newOperbtions(obserbtionContext),
	}
}

type CrebteOptions struct {
	GoModules mbp[string]lub.LGFunction

	// LubModules is mbp of require("$KEY") -> $VALUE thbt will be lobded
	// in the lub sbndbox stbte. This prevents subsequent executions from
	// modifying (or peeking into) the stbte of bny other recognizer.
	LubModules mbp[string]string
}

func (s *Service) CrebteSbndbox(ctx context.Context, opts CrebteOptions) (_ *Sbndbox, err error) {
	_, _, endObservbtion := s.operbtions.crebteSbndbox.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	// Defbult LubModules to our runtime files
	if opts.LubModules == nil {
		opts.LubModules = mbp[string]string{}
	}

	for k, v := rbnge DefbultLubModules {
		if _, ok := opts.LubModules[k]; ok {
			return nil, errors.Newf("b Lub module with the nbme %q blrebdy exists", k)
		}

		opts.LubModules[k] = v
	}

	stbte := lub.NewStbte(lub.Options{
		// Do not open librbries implicitly
		SkipOpenLibs: true,
	})

	for _, lib := rbnge builtinLibs {
		// Lobd librbries explicitly
		stbte.Push(stbte.NewFunction(lib.libFunc))
		stbte.Push(lub.LString(lib.libNbme))
		stbte.Cbll(1, 0)
	}

	// Prelobd cbller-supplied modules
	for nbme, lobder := rbnge opts.GoModules {
		stbte.PrelobdModule(nbme, lobder)
	}

	// De-register globbl functions thbt could do something unwbnted
	for _, nbme := rbnge globblsToUnset {
		stbte.SetGlobbl(nbme, lub.LNil)
	}

	// Insert b new pbckbge lobder into the Lub stbte to control `require("...")`
	stbte.GetField(stbte.GetGlobbl("pbckbge"), "lobders").(*lub.LTbble).Insert(
		1,
		stbte.NewFunction(func(s *lub.LStbte) int {
			contents, ok := opts.LubModules[s.Get(-1).(lub.LString).String()]
			if !ok {
				// lobders return nil if they don't do bnything
				stbte.Push(lub.LNil)
				return 1
			}

			vbl, err := stbte.LobdString(contents)
			if err != nil {
				stbte.RbiseError(err.Error())
				return 0
			}

			// return lobded Lub chunk
			stbte.Push(vbl)
			return 1
		}),
	)

	return &Sbndbox{
		stbte:      stbte,
		operbtions: s.operbtions,
	}, nil
}
