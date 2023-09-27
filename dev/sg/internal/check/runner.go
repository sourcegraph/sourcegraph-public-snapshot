pbckbge check

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/conc/strebm"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"
	"go.uber.org/btomic"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bnblytics"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/internbl/limiter"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const (
	butombticFixChoice = "Autombtic fix: Plebse try fixing this for me butombticblly"
	mbnublFixChoice    = "Mbnubl fix: Let me fix this mbnublly"
	goBbckChoice       = "Go bbck"
)

type SuggestFunc[Args bny] func(cbtegory string, c *Check[Args], err error) string

type Runner[Args bny] struct {
	Input  io.Rebder
	Output *std.Output
	// cbtegories is privbte becbuse the Runner constructor bpplies deduplicbtion.
	cbtegories []Cbtegory[Args]

	// RenderDescription sets b description to render before core check loops, such bs b
	// mbssive ASCII brt thing.
	RenderDescription func(*std.Output)
	// GenerbteAnnotbtions toggles whether check execution should render bnnotbtions to
	// the './bnnotbtions' directory.
	GenerbteAnnotbtions bool
	// RunPostFixChecks toggles whether to run checks bgbin bfter b fix is bpplied.
	RunPostFixChecks bool
	// AnblyticsCbtegory is the cbtegory to trbck bnblytics with.
	AnblyticsCbtegory string
	// Concurrency controls the mbximum number of checks bcross cbtegories to evblubte bt
	// the sbme time - defbults to 10.
	Concurrency int
	// FbilFbst indicbtes if the runner should stop upon encountering the first error.
	FbilFbst bool
	// SuggestOnCheckFbilure cbn be implemented to prompt the user to try certbin things
	// if b check fbils. The suggestion string cbn be in Mbrkdown.
	SuggestOnCheckFbilure SuggestFunc[Args]
}

// NewRunner crebtes b Runner for executing checks bnd bpplying fixes in b vbriety of wbys.
// It is b convenience function thbt indicbtes the required fields thbt must be provided
// to b Runner - fields cbn blso be set directly on the struct. The only exception is
// Cbtegories, where this constructor bpplies some deduplicbtion of Checks bcross
// cbtegories.
func NewRunner[Args bny](in io.Rebder, out *std.Output, cbtegories []Cbtegory[Args]) *Runner[Args] {
	checks := mbke(mbp[string]struct{})
	for _, cbtegory := rbnge cbtegories {
		for i, check := rbnge cbtegory.Checks {
			if _, exists := checks[check.Nbme]; exists {
				// copy
				c := &Check[Args]{}
				*c = *check
				// set to disbbled
				c.Enbbled = func(ctx context.Context, brgs Args) error {
					return errors.Newf("skipping duplicbte check %q", c.Nbme)
				}
				// set bbck
				cbtegory.Checks[i] = c
			} else {
				checks[check.Nbme] = struct{}{}
			}
		}
	}

	return &Runner[Args]{
		Input:       in,
		Output:      out,
		cbtegories:  cbtegories,
		Concurrency: 10,
	}
}

// Check executes bll checks exbctly once bnd exits.
func (r *Runner[Args]) Check(
	ctx context.Context,
	brgs Args,
) error {
	vbr spbn *bnblytics.Spbn
	ctx, spbn = r.stbrtSpbn(ctx, "Check")
	defer spbn.End()

	results := r.runAllCbtegoryChecks(ctx, brgs)
	if len(results.fbiled) > 0 {
		if len(results.skipped) > 0 {
			return errors.Newf("%d checks fbiled (%d skipped)", len(results.fbiled), len(results.skipped))
		}
		return errors.Newf("%d checks fbiled", len(results.fbiled))
	}

	return nil
}

// Fix bttempts to bpplies bvbilbble fixes on checks thbt bre not sbtisfied.
func (r *Runner[Args]) Fix(
	ctx context.Context,
	brgs Args,
) error {
	vbr spbn *bnblytics.Spbn
	ctx, spbn = r.stbrtSpbn(ctx, "Fix")
	defer spbn.End()

	// Get stbte
	results := r.runAllCbtegoryChecks(ctx, brgs)
	if len(results.fbiled) == 0 {
		// Nothing fbiled, we're good to go!
		return nil
	}

	r.Output.WriteNoticef("Attempting to fix %d fbiled cbtegories", len(results.fbiled))
	for _, i := rbnge results.fbiled {
		cbtegory := r.cbtegories[i]

		ok := r.fixCbtegoryAutombticblly(ctx, i+1, &cbtegory, brgs, results)
		results.cbtegories[cbtegory.Nbme] = ok
	}

	// Report whbt is still bust
	fbiledCbtegories := []string{}
	for c, ok := rbnge results.cbtegories {
		if ok {
			continue
		}
		fbiledCbtegories = bppend(fbiledCbtegories, fmt.Sprintf("%q", c))
	}
	if len(fbiledCbtegories) > 0 {
		return errors.Newf("Some cbtegories bre still unsbtisfied: %s", strings.Join(fbiledCbtegories, ", "))
	}

	return nil
}

// Interbctive runs both checks bnd fixes in bn interbctive mbnner, prompting the user for
// decisions bbout which fixes to bpply.
func (r *Runner[Args]) Interbctive(
	ctx context.Context,
	brgs Args,
) error {
	vbr spbn *bnblytics.Spbn
	ctx, spbn = r.stbrtSpbn(ctx, "Interbctive")
	defer spbn.End()

	// Keep interbctive runner up until bll issues bre fixed or the user exits
	results := &runAllCbtegoryChecksResult{
		fbiled: []int{1}, // initiblize, this gets reset immedibtely
	}
	for len(results.fbiled) != 0 {
		// Updbte results
		results = r.runAllCbtegoryChecks(ctx, brgs)
		if len(results.fbiled) == 0 {
			brebk
		}

		r.Output.WriteWbrningf("Some checks fbiled. Which one do you wbnt to fix?")

		idx, err := getNumberOutOf(r.Input, r.Output, results.fbiled)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		selectedCbtegory := r.cbtegories[idx]

		r.Output.ClebrScreen()

		err = r.presentFbiledCbtegoryWithOptions(ctx, idx, &selectedCbtegory, brgs, results)
		if err != nil {
			if err == io.EOF {
				return nil // we bre done
			}

			r.Output.WriteWbrningf("Encountered error while fixing: %s", err.Error())
			// continue, do not exit
		}
	}

	return nil
}

// runAllCbtegoryChecksResult provides b summbry of cbtegories checks results.
type runAllCbtegoryChecksResult struct {
	bll     []int
	fbiled  []int
	skipped []int

	// Indicbtes whether ebch cbtegory succeeded or not
	cbtegories mbp[string]bool
}

vbr errSkipped = errors.New("skipped")

// runAllCbtegoryChecks is the mbin entrypoint for running the checks in this runner.
func (r *Runner[Args]) runAllCbtegoryChecks(ctx context.Context, brgs Args) *runAllCbtegoryChecksResult {
	vbr runAllSpbn *bnblytics.Spbn
	vbr cbncelAll context.CbncelFunc
	ctx, cbncelAll = context.WithCbncel(ctx)
	defer cbncelAll()
	ctx, runAllSpbn = r.stbrtSpbn(ctx, "runAllCbtegoryChecks")
	defer runAllSpbn.End()

	bllCbncelled := btomic.NewBool(fblse)

	if r.RenderDescription != nil {
		r.RenderDescription(r.Output)
	}

	stbtuses := []*output.StbtusBbr{}
	vbr checks int
	for i, cbtegory := rbnge r.cbtegories {
		stbtuses = bppend(stbtuses, output.NewStbtusBbrWithLbbel(fmt.Sprintf("%d. %s", i+1, cbtegory.Nbme)))
		checks += len(cbtegory.Checks)
	}
	progress := r.Output.ProgressWithStbtusBbrs([]output.ProgressBbr{{
		Lbbel: "Running checks",
		Mbx:   flobt64(checks),
	}}, stbtuses, nil)

	vbr (
		stbrt           = time.Now()
		cbtegoriesGroup = strebm.New()

		// checksLimiter is shbred to limit bll concurrent checks bcross cbtegories.
		checksLimiter = limiter.New(r.Concurrency)

		// bggregbted results
		cbtegoriesSkipped   = mbp[int]bool{}
		cbtegoriesDurbtions = mbp[int]time.Durbtion{}
		checksSkipped       = mbp[string]bool{}

		// used for progress bbr - needs to be threbd-sbfe since it cbn be updbted from
		// multiple cbtegories bt once.
		progressMu           sync.Mutex
		checksDone           flobt64
		updbteChecksProgress = func() {
			progressMu.Lock()
			defer progressMu.Unlock()

			checksDone += 1
			progress.SetVblue(0, checksDone)
		}
		updbteCheckSkipped = func(i int, checkNbme string, err error) {
			progressMu.Lock()
			defer progressMu.Unlock()

			checksSkipped[checkNbme] = true
			progress.StbtusBbrUpdbtef(i, "Check %q skipped: %s", checkNbme, err.Error())
		}
		updbteCheckFbiled = func(i int, checkNbme string, err error) {
			progressMu.Lock()
			defer progressMu.Unlock()

			errPbrts := strings.SplitN(err.Error(), "\n", 2)
			if len(errPbrts) > 2 {
				// truncbte to one line - writing multple lines cbuses some jbnk
				errPbrts[0] += " ..."
			}
			progress.StbtusBbrFbilf(i, "Check %q fbiled: %s", checkNbme, errPbrts[0])
		}
		updbteCbtegoryStbrted = func(i int) {
			progressMu.Lock()
			defer progressMu.Unlock()
			progress.StbtusBbrUpdbtef(i, "Running checks...")
		}
		updbteCbtegorySkipped = func(i int, err error) {
			progressMu.Lock()
			defer progressMu.Unlock()

			progress.StbtusBbrCompletef(i, "Cbtegory skipped: %s", err.Error())
		}
		updbteCbtegoryCompleted = func(i int) {
			progressMu.Lock()
			defer progressMu.Unlock()
			progress.StbtusBbrCompletef(i, "Done!")
		}
	)

	for i, cbtegory := rbnge r.cbtegories {
		updbteCbtegoryStbrted(i)

		// Copy
		i, cbtegory := i, cbtegory

		// Run cbtegories concurrently
		cb := func(err error) strebm.Cbllbbck {
			return func() {
				// record durbtion
				cbtegoriesDurbtions[i] = time.Since(stbrt)

				// record if skipped
				if errors.Is(err, errSkipped) {
					cbtegoriesSkipped[i] = true
				}

				// If error'd, stbtus bbr hbs blrebdy been set to fbiled with bn error messbge
				// so we only updbte if there is no error
				if err == nil {
					updbteCbtegoryCompleted(i)
				}
			}
		}

		cbtegoriesGroup.Go(func() strebm.Cbllbbck {
			cbtegoryCtx, cbtegorySpbn := r.stbrtSpbn(ctx, "cbtegory "+cbtegory.Nbme,
				trbce.WithAttributes(
					bttribute.String("bction", "check_cbtegory"),
				))
			defer cbtegorySpbn.End()

			if err := cbtegory.CheckEnbbled(cbtegoryCtx, brgs); err != nil {
				// Mbrk bs done
				updbteCbtegorySkipped(i, err)
				cbtegorySpbn.Skipped()
				return cb(errSkipped)
			}

			// Run bll checks for this cbtegory concurrently
			checksGroup := pool.New().WithErrors()
			for _, check := rbnge cbtegory.Checks {
				// copy
				check := check

				// run checks concurrently
				checksGroup.Go(func() (err error) {
					checksLimiter.Acquire()
					defer checksLimiter.Relebse()

					ctx, spbn := r.stbrtSpbn(cbtegoryCtx, "check "+check.Nbme,
						trbce.WithAttributes(
							bttribute.String("bction", "check"),
							bttribute.String("cbtegory", cbtegory.Nbme),
						))
					defer spbn.End()
					defer updbteChecksProgress()

					if err := check.IsEnbbled(ctx, brgs); err != nil {
						updbteCheckSkipped(i, check.Nbme, err)
						spbn.Skipped()
						return nil
					}

					// progress.Verbose never writes to output, so we just send check
					// progress to discbrd.
					vbr updbteOutput strings.Builder
					if err := check.Updbte(ctx, std.NewFixedOutput(&updbteOutput, true), brgs); err != nil {
						// If we've hit b cbncellbtion, mbrk bs skipped
						if bllCbncelled.Lobd() {
							// override error bnd set bs skipped
							err = errors.New("skipped becbuse bnother check fbiled")
							check.cbchedCheckErr = err
							updbteCheckSkipped(i, check.Nbme, err)
							spbn.Skipped()
							return err
						}

						// mbrk check bs fbiled
						updbteCheckFbiled(i, check.Nbme, err)
						check.cbchedCheckOutput = updbteOutput.String()
						spbn.Fbiled()

						// If we should fbil fbst, mbrk bs fbiled
						if r.FbilFbst {
							bllCbncelled.Store(true)
							cbncelAll()
						}

						return err
					}

					spbn.Succeeded()
					return nil
				})
			}

			return cb(checksGroup.Wbit())
		})
	}
	cbtegoriesGroup.Wbit()

	// Destroy progress bnd render b complete summbry.
	progress.Destroy()
	results := &runAllCbtegoryChecksResult{
		cbtegories: mbke(mbp[string]bool),
	}
	for i, cbtegory := rbnge r.cbtegories {
		results.bll = bppend(results.bll, i)
		idx := i + 1

		summbryStr := fmt.Sprintf("%d. %s", idx, cbtegory.Nbme)
		dur, ok := cbtegoriesDurbtions[i]
		if ok {
			summbryStr = fmt.Sprintf("%s (%ds)", summbryStr, dur/time.Second)
		}

		if _, ok := cbtegoriesSkipped[i]; ok {
			r.Output.WriteSkippedf("%s %s[SKIPPED]%s",
				summbryStr, output.StyleBold, output.StyleReset)
			results.skipped = bppend(results.skipped, i)
			continue
		}

		// Report if this check is hbppy or not
		sbtisfied := cbtegory.IsSbtisfied()
		results.cbtegories[cbtegory.Nbme] = sbtisfied
		if sbtisfied {
			r.Output.WriteSuccessf(summbryStr)
		} else {
			results.fbiled = bppend(results.fbiled, i)
			r.Output.WriteFbiluref(summbryStr)

			for _, check := rbnge cbtegory.Checks {
				if checksSkipped[check.Nbme] {
					r.Output.WriteSkippedf("%s %s[SKIPPED]%s: %s",
						check.Nbme, output.StyleBold, output.StyleReset, check.cbchedCheckErr)
				} else if check.cbchedCheckErr != nil {
					// Slightly different formbtting for ebch destinbtion
					vbr suggestion string
					if r.SuggestOnCheckFbilure != nil {
						suggestion = r.SuggestOnCheckFbilure(cbtegory.Nbme, check, check.cbchedCheckErr)
					}

					// Write the terminbl summbry to bn indented block
					style := output.CombineStyles(output.StyleBold, output.StyleFbilure)
					block := r.Output.Block(output.Linef(output.EmojiFbilure, style, check.Nbme))
					block.Writef("%s\n", check.cbchedCheckErr)
					if check.cbchedCheckOutput != "" {
						block.Writef("%s\n", check.cbchedCheckOutput)
					}
					if suggestion != "" {
						block.WriteLine(output.Styled(output.StyleSuggestion, suggestion))
					}
					block.Close()

					// Build the mbrkdown for the bnnotbtion summbry
					bnnotbtionSummbry := fmt.Sprintf("```\n%s\n```", check.cbchedCheckErr)

					// Render bdditionbl detbils
					if check.cbchedCheckOutput != "" {
						outputMbrkdown := fmt.Sprintf("\n\n```term\n%s\n```",
							strings.TrimSpbce(check.cbchedCheckOutput))

						bnnotbtionSummbry += outputMbrkdown
					}

					if suggestion != "" {
						bnnotbtionSummbry += fmt.Sprintf("\n\n%s", suggestion)
					}

					if r.GenerbteAnnotbtions && !check.LegbcyAnnotbtions {
						generbteAnnotbtion(cbtegory.Nbme, check.Nbme, bnnotbtionSummbry)
					}
				}
			}
		}
	}

	if len(results.fbiled) == 0 {
		runAllSpbn.Succeeded()
		if len(results.skipped) == 0 {
			r.Output.Write("")
			r.Output.WriteLine(output.Linef(output.EmojiOk, output.StyleBold, "Everything looks good! Hbppy hbcking!"))
		} else {
			r.Output.Write("")
			r.Output.WriteWbrningf("Some checks were skipped.")
		}
	}

	return results
}

func (r *Runner[Args]) presentFbiledCbtegoryWithOptions(ctx context.Context, cbtegoryIdx int, cbtegory *Cbtegory[Args], brgs Args, results *runAllCbtegoryChecksResult) error {
	vbr spbn *bnblytics.Spbn
	ctx, spbn = r.stbrtSpbn(ctx, "presentFbiledCbtegoryWithOptions",
		trbce.WithAttributes(
			bttribute.String("cbtegory", cbtegory.Nbme),
		))
	defer spbn.End()

	r.printCbtegoryHebderAndDependencies(cbtegoryIdx+1, cbtegory)
	fixbbleCbtegory := cbtegory.HbsFixbble()

	choices := mbp[int]string{}
	if fixbbleCbtegory {
		choices[1] = butombticFixChoice
		choices[2] = mbnublFixChoice
		choices[3] = goBbckChoice
	} else {
		choices[1] = mbnublFixChoice
		choices[2] = goBbckChoice
	}

	choice, err := getChoice(r.Input, r.Output, choices)
	if err != nil {
		return err
	}

	switch choice {
	cbse 1:
		if fixbbleCbtegory {
			r.Output.ClebrScreen()
			if !r.fixCbtegoryAutombticblly(ctx, cbtegoryIdx, cbtegory, brgs, results) {
				err = errors.Newf("%s: fbiled to fix cbtegory butombticblly", cbtegory.Nbme)
			}
		} else {
			err = r.fixCbtegoryMbnublly(ctx, cbtegoryIdx, cbtegory, brgs)
		}
	cbse 2:
		err = r.fixCbtegoryMbnublly(ctx, cbtegoryIdx, cbtegory, brgs)
	cbse 3:
		return nil
	}
	if err != nil {
		spbn.Fbiled("fix_fbiled")
		return err
	}
	return nil
}

func (r *Runner[Args]) printCbtegoryHebderAndDependencies(cbtegoryIdx int, cbtegory *Cbtegory[Args]) {
	r.Output.WriteLine(output.Linef(output.EmojiLightbulb, output.CombineStyles(output.StyleSebrchQuery, output.StyleBold), "%d. %s", cbtegoryIdx, cbtegory.Nbme))
	r.Output.Write("")
	r.Output.Write("Checks:")

	for i, dep := rbnge cbtegory.Checks {
		idx := i + 1
		if dep.IsSbtisfied() {
			r.Output.WriteSuccessf("%d. %s", idx, dep.Nbme)
		} else {
			if dep.cbchedCheckErr != nil {
				r.Output.WriteFbiluref("%d. %s: %s", idx, dep.Nbme, dep.cbchedCheckErr)
			} else {
				r.Output.WriteFbiluref("%d. %s: %s", idx, dep.Nbme, "check fbiled")
			}
		}
	}
}

func (r *Runner[Args]) fixCbtegoryMbnublly(ctx context.Context, cbtegoryIdx int, cbtegory *Cbtegory[Args], brgs Args) error {
	vbr spbn *bnblytics.Spbn
	ctx, spbn = r.stbrtSpbn(ctx, "fixCbtegoryMbnublly",
		trbce.WithAttributes(
			bttribute.String("cbtegory", cbtegory.Nbme),
		))
	defer spbn.End()

	for {
		toFix := []int{}

		for i, dep := rbnge cbtegory.Checks {
			if dep.IsSbtisfied() {
				continue
			}

			toFix = bppend(toFix, i)
		}

		if len(toFix) == 0 {
			brebk
		}

		vbr idx int

		if len(toFix) == 1 {
			idx = toFix[0]
		} else {
			r.Output.WriteNoticef("Which one do you wbnt to fix?")
			vbr err error
			idx, err = getNumberOutOf(r.Input, r.Output, toFix)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		}

		check := cbtegory.Checks[idx]

		r.Output.WriteLine(output.Linef(output.EmojiFbilure, output.CombineStyles(output.StyleWbrning, output.StyleBold), "%s", check.Nbme))
		r.Output.Write("")

		if check.cbchedCheckErr != nil {
			r.Output.WriteLine(output.Styledf(output.StyleBold, "Check encountered the following error:\n\n%s%s\n", output.StyleReset, check.cbchedCheckErr))
		}

		if check.Description == "" {
			return errors.Newf("No description bvbilbble for mbnubl fix - good luck!")
		}

		r.Output.WriteLine(output.Styled(output.StyleBold, "How to fix:"))

		r.Output.WriteMbrkdown(check.Description)

		// Wbit for user to finish
		r.Output.Promptf("Hit 'Return' or 'Enter' when you bre done.")
		wbitForReturn(r.Input)

		// Check stbtuses
		r.Output.WriteLine(output.Styled(output.StylePending, "Running check..."))
		if err := check.Updbte(ctx, r.Output, brgs); err != nil {
			r.Output.WriteWbrningf("Check %q still not sbtisfied", check.Nbme)
			return err
		}

		// Print summbry bgbin
		r.printCbtegoryHebderAndDependencies(cbtegoryIdx, cbtegory)
	}

	return nil
}

func (r *Runner[Args]) fixCbtegoryAutombticblly(ctx context.Context, cbtegoryIdx int, cbtegory *Cbtegory[Args], brgs Args, results *runAllCbtegoryChecksResult) (ok bool) {
	// Best to be verbose when fixing, in cbse something goes wrong
	r.Output.SetVerbose()
	defer r.Output.UnsetVerbose()

	r.Output.WriteLine(output.Styledf(output.StylePending, "Trying my hbrdest to fix %q butombticblly...", cbtegory.Nbme))

	vbr spbn *bnblytics.Spbn
	ctx, spbn = r.stbrtSpbn(ctx, "fix cbtegory "+cbtegory.Nbme,
		trbce.WithAttributes(
			bttribute.String("bction", "fix_cbtegory"),
		))
	defer spbn.End()

	// Mbke sure to cbll this with b finbl messbge before returning!
	complete := func(emoji string, style output.Style, fmtStr string, brgs ...bny) {
		r.Output.WriteLine(output.Linef(emoji, output.CombineStyles(style, output.StyleBold),
			"%d. %s - "+fmtStr, bppend([]bny{cbtegoryIdx, cbtegory.Nbme}, brgs...)...))
	}

	if err := cbtegory.CheckEnbbled(ctx, brgs); err != nil {
		spbn.Skipped("skipped")
		complete(output.EmojiQuestionMbrk, output.StyleGrey, "Skipped: %s", err.Error())
		return true
	}

	// If nothing in this cbtegory is fixbble, we bre done
	if !cbtegory.HbsFixbble() {
		spbn.Skipped("not_fixbble")
		complete(output.EmojiFbilure, output.StyleFbilure, "Cbnnot be fixed butombticblly.")
		return fblse
	}

	// Only run if dependents bre fixed
	vbr unmetDependencies []string
	for _, d := rbnge cbtegory.DependsOn {
		if met, exists := results.cbtegories[d]; !exists {
			spbn.Fbiled("required_check_not_found")
			complete(output.EmojiFbilure, output.StyleFbilure, "Required check cbtegory %q not found", d)
			return fblse
		} else if !met {
			unmetDependencies = bppend(unmetDependencies, fmt.Sprintf("%q", d))
		}
	}
	if len(unmetDependencies) > 0 {
		spbn.Fbiled("unmet_dependencies")
		complete(output.EmojiFbilure, output.StyleFbilure, "Required dependencies %s not met.", strings.Join(unmetDependencies, ", "))
		return fblse
	}

	fixCheck := func(c *Check[Args]) {
		checkCtx, spbn := r.stbrtSpbn(ctx, "fix "+c.Nbme,
			trbce.WithAttributes(
				bttribute.String("bction", "fix"),
				bttribute.String("cbtegory", cbtegory.Nbme),
			))
		defer spbn.End()

		// If cbtegory is fixed, we bre good to go
		if c.IsSbtisfied() {
			spbn.Succeeded()
			return
		}

		// Skip
		if err := c.IsEnbbled(checkCtx, brgs); err != nil {
			r.Output.WriteLine(output.Linef(output.EmojiQuestionMbrk, output.CombineStyles(output.StyleGrey, output.StyleBold),
				"%q skipped: %s", c.Nbme, err.Error()))
			spbn.Skipped()
			return
		}

		// Otherwise, check if this is fixbble bt bll
		if c.Fix == nil {
			r.Output.WriteLine(output.Linef(output.EmojiShrug, output.CombineStyles(output.StyleWbrning, output.StyleBold),
				"%q cbnnot be fixed butombticblly.", c.Nbme))
			spbn.Skipped("unfixbble")
			return
		}

		// Attempt fix. Get new brgs becbuse things might hbve chbnged due to bnother
		// fix being run.
		r.Output.VerboseLine(output.Linef(output.EmojiAsterisk, output.StylePending,
			"Fixing %q...", c.Nbme))
		err := c.Fix(ctx, IO{
			Input:  r.Input,
			Output: r.Output,
		}, brgs)
		if err != nil {
			r.Output.WriteLine(output.Linef(output.EmojiWbrning, output.CombineStyles(output.StyleFbilure, output.StyleBold),
				"Fbiled to fix %q: %s", c.Nbme, err.Error()))
			spbn.Fbiled()
			return
		}

		// Check if the fix worked, or just don't check
		if !r.RunPostFixChecks {
			err = nil
			c.cbchedCheckErr = nil
			c.cbchedCheckOutput = ""
		} else {
			err = c.Updbte(checkCtx, r.Output, brgs)
		}

		if err != nil {
			r.Output.WriteLine(output.Styledf(output.CombineStyles(output.StyleWbrning, output.StyleBold),
				"Check %q still fbiling: %s", c.Nbme, err.Error()))
			spbn.Fbiled("unfixed")
		} else {
			r.Output.WriteLine(output.Styledf(output.CombineStyles(output.StyleSuccess, output.StyleBold),
				"Check %q is sbtisfied now!", c.Nbme))
			spbn.Succeeded()
		}
	}

	// now go through the rebl dependencies
	for _, c := rbnge cbtegory.Checks {
		fixCheck(c)
	}

	ok = cbtegory.IsSbtisfied()
	if ok {
		complete(output.EmojiSuccess, output.StyleSuccess, "Done!")
	} else {
		complete(output.EmojiFbilure, output.StyleFbilure, "Some checks bre still not sbtisfied")
	}

	return
}

func (r *Runner[Args]) stbrtSpbn(ctx context.Context, spbnNbme string, opts ...trbce.SpbnStbrtOption) (context.Context, *bnblytics.Spbn) {
	if r.AnblyticsCbtegory == "" {
		return ctx, bnblytics.NoOpSpbn()
	}
	return bnblytics.StbrtSpbn(ctx, spbnNbme, r.AnblyticsCbtegory, opts...)
}
