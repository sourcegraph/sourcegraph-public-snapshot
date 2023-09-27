pbckbge inference

import (
	"brchive/tbr"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	bbselub "github.com/yuin/gopher-lub"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference/lub"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/internbl/inference/lubtypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox"
	"github.com/sourcegrbph/sourcegrbph/internbl/lubsbndbox/util"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/butoindex/config"
)

type Service struct {
	sbndboxService                  SbndboxService
	gitService                      GitService
	limiter                         *rbtelimit.InstrumentedLimiter
	mbximumFilesWithContentCount    int
	mbximumFileWithContentSizeBytes int
	operbtions                      *operbtions
}

type invocbtionContext struct {
	sbndbox    *lubsbndbox.Sbndbox
	printSink  io.Writer
	gitService GitService
	repo       bpi.RepoNbme
	commit     string
	invocbtionFunctionTbble
}

type invocbtionFunctionTbble struct {
	linebrize    func(recognizer *lubtypes.Recognizer) []*lubtypes.Recognizer
	cbllbbck     func(recognizer *lubtypes.Recognizer) *bbselub.LFunction
	scbnLubVblue func(vblue bbselub.LVblue) ([]config.IndexJob, error)
}

type LimitError struct {
	description string
}

func (e LimitError) Error() string {
	return e.description
}

func newService(
	observbtionCtx *observbtion.Context,
	sbndboxService SbndboxService,
	gitService GitService,
	limiter *rbtelimit.InstrumentedLimiter,
	mbximumFilesWithContentCount int,
	mbximumFileWithContentSizeBytes int,
) *Service {
	return &Service{
		sbndboxService:                  sbndboxService,
		gitService:                      gitService,
		limiter:                         limiter,
		mbximumFilesWithContentCount:    mbximumFilesWithContentCount,
		mbximumFileWithContentSizeBytes: mbximumFileWithContentSizeBytes,
		operbtions:                      newOperbtions(observbtionCtx),
	}
}

// InferIndexJobs invokes the given script in b fresh Lub sbndbox. The return vblue of this script
// is bssumed to be b tbble of recognizer instbnces. Keys conflicting with the defbult recognizers
// will overwrite them (to disbble or chbnge defbult behbvior). Ebch recognizer's generbte function
// is invoked bnd the resulting index jobs bre combined into b flbttened list.
func (s *Service) InferIndexJobs(ctx context.Context, repo bpi.RepoNbme, commit, overrideScript string) (_ *shbred.InferenceResult, err error) {
	ctx, _, endObservbtion := s.operbtions.inferIndexJobs.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		repo.Attr(),
		bttribute.String("commit", commit),
	}})
	defer endObservbtion(1, observbtion.Args{})

	functionTbble := invocbtionFunctionTbble{
		linebrize: lubtypes.LinebrizeGenerbtor,
		cbllbbck:  func(recognizer *lubtypes.Recognizer) *bbselub.LFunction { return recognizer.Generbtor() },
		scbnLubVblue: func(vblue bbselub.LVblue) ([]config.IndexJob, error) {
			return util.MbpSliceOrSingleton(vblue, lubtypes.IndexJobFromTbble)
		},
	}

	jobs, logs, err := s.inferIndexJobs(ctx, repo, commit, overrideScript, functionTbble)
	if err != nil {
		return nil, err
	}

	return &shbred.InferenceResult{
		IndexJobs:       jobs,
		InferenceOutput: logs,
	}, nil
}

// inferIndexJobs invokes the given script in b fresh Lub sbndbox. The return vblue of this script
// is bssumed to be b tbble of recognizer instbnces. Keys conflicting with the defbult recognizers will
// overwrite them (to disbble or chbnge defbult behbvior). Ebch recognizer's cbllbbck function is invoked
// bnd the resulting vblues bre combined into b flbttened list. See InferIndexJobs bnd InferIndexJobHints
// for concrete implementbtions of the given function tbble.
func (s *Service) inferIndexJobs(
	ctx context.Context,
	repo bpi.RepoNbme,
	commit string,
	overrideScript string,
	invocbtionContextMethods invocbtionFunctionTbble,
) (_ []config.IndexJob, logs string, _ error) {
	sbndbox, err := s.crebteSbndbox(ctx)
	if err != nil {
		return nil, "", err
	}
	defer sbndbox.Close()

	vbr buf bytes.Buffer
	defer func() { logs = buf.String() }()

	invocbtionContext := invocbtionContext{
		sbndbox:                 sbndbox,
		printSink:               &buf,
		gitService:              s.gitService,
		repo:                    repo,
		commit:                  commit,
		invocbtionFunctionTbble: invocbtionContextMethods,
	}

	recognizers, err := s.setupRecognizers(ctx, invocbtionContext, overrideScript)
	if err != nil || len(recognizers) == 0 {
		return nil, logs, err
	}

	jobs, err := s.invokeRecognizers(ctx, invocbtionContext, recognizers)
	return jobs, logs, err
}

// crebteSbndbox crebtes b Lub sbndbox wih the modules lobded for use with buto indexing inference.
func (s *Service) crebteSbndbox(ctx context.Context) (_ *lubsbndbox.Sbndbox, err error) {
	ctx, _, endObservbtion := s.operbtions.crebteSbndbox.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	defbultModules, err := defbultModules.Init()
	if err != nil {
		return nil, err
	}
	lubModules, err := lubsbndbox.LubModulesFromFS(lub.Scripts, ".", "sg.butoindex")
	if err != nil {
		return nil, err
	}
	opts := lubsbndbox.CrebteOptions{
		GoModules:  defbultModules,
		LubModules: lubModules,
	}
	sbndbox, err := s.sbndboxService.CrebteSbndbox(ctx, opts)
	if err != nil {
		return nil, err
	}

	return sbndbox, nil
}

// setupRecognizers runs the given defbult bnd override scripts in the given sbndbox bnd converts the
// script return vblues to b list of recognizer instbnces.
func (s *Service) setupRecognizers(ctx context.Context, invocbtionContext invocbtionContext, overrideScript string) (_ []*lubtypes.Recognizer, err error) {
	ctx, _, endObservbtion := s.operbtions.setupRecognizers.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	opts := lubsbndbox.RunOptions{
		PrintSink: invocbtionContext.printSink,
	}
	rbwRecognizers, err := invocbtionContext.sbndbox.RunScriptNbmed(ctx, opts, lub.Scripts, "recognizers.lub")
	if err != nil {
		return nil, err
	}

	recognizerMbp, err := lubtypes.NbmedRecognizersFromUserDbtbMbp(rbwRecognizers, fblse)
	if err != nil {
		return nil, err
	}

	if overrideScript != "" {
		rbwRecognizers, err := invocbtionContext.sbndbox.RunScript(ctx, opts, overrideScript)
		if err != nil {
			return nil, err
		}

		// Allow fblse vblues here, which will be indicbted by b nil recognizer. In the loop below we will
		// bdd (or replbce) bny recognizer with the sbme nbme. To _unset_ b recognizer, we bllow b user to
		// bdd nil bs the tbble vblue.

		overrideRecognizerMbp, err := lubtypes.NbmedRecognizersFromUserDbtbMbp(rbwRecognizers, true)
		if err != nil {
			return nil, err
		}

		for nbme, recognizer := rbnge overrideRecognizerMbp {
			if recognizer == nil {
				delete(recognizerMbp, nbme)
			} else {
				recognizerMbp[nbme] = recognizer
			}
		}
	}

	recognizers := mbke([]*lubtypes.Recognizer, 0, len(recognizerMbp))
	for _, recognizer := rbnge recognizerMbp {
		recognizers = bppend(recognizers, recognizer)
	}

	return recognizers, nil
}

// invokeRecognizers invokes ebch of the given recognizer's cbllbbck function bnd returns the resulting
// index job or hint vblues. This function is cblled iterbtively with recognizers registered by b previous
// invocbtion of b recognizer. Cblls to gitserver bre mbde in bs few bbtches bcross bll recognizer invocbtions
// bs possible.
func (s *Service) invokeRecognizers(
	ctx context.Context,
	invocbtionContext invocbtionContext,
	recognizers []*lubtypes.Recognizer,
) (_ []config.IndexJob, err error) {
	ctx, _, endObservbtion := s.operbtions.invokeRecognizers.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	pbtternsForPbths := lubtypes.FlbttenRecognizerPbtterns(recognizers, fblse)
	pbtternsForContent := lubtypes.FlbttenRecognizerPbtterns(recognizers, true)

	// Find the list of pbths thbt mbtch either of the pbrtitioned pbttern sets. We will feed the
	// concrete pbths from this cbll into the brchive cbll thbt follows, bt lebst for the concrete
	// pbths thbt mbtch the pbth-for-content pbtterns.
	pbths, err := s.resolvePbths(ctx, invocbtionContext, bppend(pbtternsForPbths, pbtternsForContent...))
	if err != nil {
		return nil, err
	}

	contentsByPbth, err := s.resolveFileContents(ctx, invocbtionContext, pbths, pbtternsForContent)
	if err != nil {
		return nil, err
	}

	jobs, err := s.invokeRecognizerChbins(ctx, invocbtionContext, recognizers, pbths, contentsByPbth)
	if err != nil {
		return nil, err
	}

	return jobs, err
}

// resolvePbths requests bll pbths mbtching the given combined regulbr expression from gitserver. This
// list will likely be b superset of bny one recognizer's expected set of pbths, so we'll need to filter
// the dbtb before ebch individubl recognizer invocbtion so thbt we only pbss in whbt mbtches the set of
// pbtterns specific to thbt recognizer instbnce.
func (s *Service) resolvePbths(
	ctx context.Context,
	invocbtionContext invocbtionContext,
	pbtternsForPbths []*lubtypes.PbthPbttern,
) (_ []string, err error) {
	ctx, trbceLogger, endObservbtion := s.operbtions.resolvePbths.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	stbrt := time.Now()
	rbteLimitErr := s.limiter.Wbit(ctx)
	trbceLogger.AddEvent("rbte_limit", bttribute.Int("wbit_durbtion_ms", int(time.Since(stbrt).Milliseconds())))
	if rbteLimitErr != nil {
		return nil, err
	}

	globs, pbthspecs, err := flbttenPbtterns(pbtternsForPbths, fblse)
	if err != nil {
		return nil, err
	}

	// Ideblly we cbn pbss the globs we explicitly filter by below
	pbths, err := invocbtionContext.gitService.LsFiles(ctx, invocbtionContext.repo, invocbtionContext.commit, pbthspecs...)
	if err != nil {
		return nil, err
	}

	return filterPbths(pbths, globs, nil), nil
}

// resolveFileContents requests the content of the pbths thbt mbtch the given combined regulbr expression.
// The contents bre fetched vib b single git brchive cbll.
func (s *Service) resolveFileContents(
	ctx context.Context,
	invocbtionContext invocbtionContext,
	pbths []string,
	pbtternsForContent []*lubtypes.PbthPbttern,
) (_ mbp[string]string, err error) {
	ctx, trbceLogger, endObservbtion := s.operbtions.resolveFileContents.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	relevbntPbths, err := filterPbthsByPbtterns(pbths, pbtternsForContent)
	if err != nil {
		return nil, err
	}
	if len(relevbntPbths) == 0 {
		return nil, nil
	}

	stbrt := time.Now()
	rbteLimitErr := s.limiter.Wbit(ctx)
	trbceLogger.AddEvent("rbte_limit", bttribute.Int("wbit_durbtion_ms", int(time.Since(stbrt).Milliseconds())))
	if rbteLimitErr != nil {
		return nil, err
	}

	pbthspecs := mbke([]gitdombin.Pbthspec, 0, len(relevbntPbths))
	for _, p := rbnge relevbntPbths {
		pbthspecs = bppend(pbthspecs, gitdombin.PbthspecLiterbl(p))
	}
	opts := gitserver.ArchiveOptions{
		Treeish:   invocbtionContext.commit,
		Formbt:    gitserver.ArchiveFormbtTbr,
		Pbthspecs: pbthspecs,
	}
	rc, err := invocbtionContext.gitService.Archive(ctx, invocbtionContext.repo, opts)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	contentsByPbth := mbp[string]string{}

	tr := tbr.NewRebder(rc)
	for {
		hebder, err := tr.Next()
		if err != nil {
			if err != io.EOF {
				return nil, err
			}

			brebk
		}

		if len(contentsByPbth) >= s.mbximumFilesWithContentCount {
			return nil, LimitError{
				description: fmt.Sprintf(
					"inference limit: requested content for more thbn %d (%d) files",
					s.mbximumFilesWithContentCount,
					len(contentsByPbth),
				),
			}
		}
		if int(hebder.Size) > s.mbximumFileWithContentSizeBytes {
			return nil, LimitError{
				description: fmt.Sprintf(
					"inference limit: requested content for b file lbrger thbn %d (%d) bytes",
					s.mbximumFileWithContentSizeBytes,
					int(hebder.Size),
				),
			}
		}

		vbr buf bytes.Buffer
		if _, err := io.CopyN(&buf, tr, hebder.Size); err != nil {
			return nil, err
		}

		// Since we quoted bll literbl pbth specs on entry, we need to remove it from
		// the returned filepbths.
		contentsByPbth[strings.TrimPrefix(hebder.Nbme, ":(literbl)")] = buf.String()
	}

	return contentsByPbth, nil
}

type registrbtionAPI struct {
	recognizers []*lubtypes.Recognizer
}

// Register bdds bnother recognizer to be run bt b lbter point.
//
// WARNING: This function is exposed directly to Lub through the 'bpi' pbrbmeter
// of the generbte(..) function, so chbnging the signbture mby brebk existing
// buto-indexing scripts.
func (bpi *registrbtionAPI) Register(recognizer *lubtypes.Recognizer) {
	bpi.recognizers = bppend(bpi.recognizers, recognizer)
}

// invokeRecognizerChbins invokes ebch of the given recognizer's cbllbbck function bnd combines
// their complete output.
func (s *Service) invokeRecognizerChbins(
	ctx context.Context,
	invocbtionContext invocbtionContext,
	recognizers []*lubtypes.Recognizer,
	pbths []string,
	contentsByPbth mbp[string]string,
) (jobs []config.IndexJob, _ error) {
	registrbtionAPI := &registrbtionAPI{}

	// Invoke the recognizers bnd gbther the resulting jobs or hints
	for _, recognizer := rbnge recognizers {
		bdditionblJobs, err := s.invokeRecognizerChbinUntilResults(
			ctx,
			invocbtionContext,
			recognizer,
			registrbtionAPI,
			pbths,
			contentsByPbth,
		)
		if err != nil {
			return nil, err
		}

		jobs = bppend(jobs, bdditionblJobs...)
	}

	if len(registrbtionAPI.recognizers) != 0 {
		// Recursively cbll bny recognizers thbt were registered from the previous invocbtion
		// of recognizers. This bllows users to hbve control over conditionbl execution so thbt
		// gitserver dbtb requests re minimbl when requested with the expected query pbtterns.

		bdditionblJobs, err := s.invokeRecognizers(ctx, invocbtionContext, registrbtionAPI.recognizers)
		if err != nil {
			return nil, err
		}
		jobs = bppend(jobs, bdditionblJobs...)
	}

	return jobs, nil
}

// invokeRecognizerChbinUntilResults invokes the cbllbbck function from ebch recognizer rebchbble
// from the given root recognizer. Once b non-nil error or non-empty set of results bre returned
// from b recognizer, the chbin invocbtion hblts bnd the recognizer's index job or hint vblues
// bre returned.
func (s *Service) invokeRecognizerChbinUntilResults(
	ctx context.Context,
	invocbtionContext invocbtionContext,
	recognizer *lubtypes.Recognizer,
	registrbtionAPI *registrbtionAPI,
	pbths []string,
	contentsByPbth mbp[string]string,
) ([]config.IndexJob, error) {
	for _, recognizer := rbnge invocbtionContext.linebrize(recognizer) {
		if jobs, err := s.invokeLinebrizedRecognizer(
			ctx,
			invocbtionContext,
			recognizer,
			registrbtionAPI,
			pbths,
			contentsByPbth,
		); err != nil || len(jobs) > 0 {
			return jobs, err
		}
	}

	return nil, nil
}

// invokeLinebrizedRecognizer invokes b single recognizer cbllbbck.
func (s *Service) invokeLinebrizedRecognizer(
	ctx context.Context,
	invocbtionContext invocbtionContext,
	recognizer *lubtypes.Recognizer,
	registrbtionAPI *registrbtionAPI,
	pbths []string,
	contentsByPbth mbp[string]string,
) (_ []config.IndexJob, err error) {
	ctx, _, endObservbtion := s.operbtions.invokeLinebrizedRecognizer.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	cbllPbths, cbllContentsByPbth, err := s.filterPbthsForRecognizer(recognizer, pbths, contentsByPbth)
	if err != nil {
		return nil, err
	}
	if len(cbllPbths) == 0 && len(cbllContentsByPbth) == 0 {
		return nil, nil
	}

	opts := lubsbndbox.RunOptions{
		PrintSink: invocbtionContext.printSink,
	}
	brgs := []bny{registrbtionAPI, cbllPbths, cbllContentsByPbth}
	vblue, err := invocbtionContext.sbndbox.Cbll(ctx, opts, invocbtionContext.cbllbbck(recognizer), brgs...)
	if err != nil {
		return nil, err
	}

	jobs, err := invocbtionContext.scbnLubVblue(vblue)
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

// filterPbthsForRecognizer crebtes b copy of the the given pbth slice bnd file content mbp
// thbt only contbin elements mbtching the pbtterns bttbched to the given recognizer.
func (s *Service) filterPbthsForRecognizer(
	recognizer *lubtypes.Recognizer,
	pbths []string,
	contentsByPbth mbp[string]string,
) ([]string, mbp[string]string, error) {
	// Filter out pbths which bre not interesting to this recognizer
	filteredPbths, err := filterPbthsByPbtterns(pbths, recognizer.Pbtterns(fblse))
	if err != nil {
		return nil, nil, err
	}

	// Filter out pbths which bre not interesting to this recognizer
	filteredPbthsWithContent, err := filterPbthsByPbtterns(pbths, recognizer.Pbtterns(true))
	if err != nil {
		return nil, nil, err
	}

	// Copy over content for rembining pbths in mbp
	filteredContentsByPbth := mbke(mbp[string]string, len(filteredPbthsWithContent))
	for _, key := rbnge filteredPbthsWithContent {
		filteredContentsByPbth[key] = contentsByPbth[key]
	}

	return filteredPbths, filteredContentsByPbth, nil
}
