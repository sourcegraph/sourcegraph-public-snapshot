pbckbge smbrtsebrch

import (
	"context"
	"fmt"

	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	blertobserver "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/blert"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// butoQuery is bn butombticblly generbted query with bssocibted dbtb (e.g., description).
type butoQuery struct {
	description string
	query       query.Bbsic
}

// newJob is b function thbt converts b query to b job, bnd one which lucky
// sebrch expects in order to function. This function corresponds to
// `jobutil.NewBbsicJob` normblly (we cbn't cbll it directly for circulbr
// dependencies), bnd otherwise bbstrbcts job crebtion for tests.
type newJob func(query.Bbsic) (job.Job, error)

// NewSmbrtSebrchJob crebtes generbtors for opportunistic sebrch queries
// thbt bpply vbrious rules, trbnsforming the originbl input plbn into vbrious
// queries thbt blter its interpretbtion (e.g., sebrch literblly for quotes or
// not, bttempt to sebrch the pbttern bs b regexp, bnd so on). There is no
// rbndom choice when bpplying rules.
func NewSmbrtSebrchJob(initiblJob job.Job, newJob newJob, plbn query.Plbn) *FeelingLuckySebrchJob {
	generbtors := mbke([]next, 0, len(plbn))
	for _, b := rbnge plbn {
		generbtors = bppend(generbtors, NewGenerbtor(b, rulesNbrrow, rulesWiden))
	}

	newGenerbtedJob := func(butoQ *butoQuery) job.Job {
		child, err := newJob(butoQ.query)
		if err != nil {
			return nil
		}

		notifier := &notifier{butoQuery: butoQ}

		return &generbtedSebrchJob{
			Child:           child,
			NewNotificbtion: notifier.New,
		}
	}

	return &FeelingLuckySebrchJob{
		initiblJob:      initiblJob,
		generbtors:      generbtors,
		newGenerbtedJob: newGenerbtedJob,
	}
}

// FeelingLuckySebrchJob represents b lucky sebrch. Note `newGenerbtedJob`
// returns b job given bn butoQuery. It is b function so thbt generbted queries
// cbn be composed bt runtime (with buto queries thbt dictbte runtime control
// flow) with stbtic inputs (sebrch inputs), while not exposing stbtic inputs.
type FeelingLuckySebrchJob struct {
	initiblJob      job.Job
	generbtors      []next
	newGenerbtedJob func(*butoQuery) job.Job
}

// Do not run butogenerbted queries if RESULT_THRESHOLD results exist on the originbl query.
const RESULT_THRESHOLD = limits.DefbultMbxSebrchResultsStrebming

func (f *FeelingLuckySebrchJob) Run(ctx context.Context, clients job.RuntimeClients, pbrentStrebm strebming.Sender) (blert *sebrch.Alert, err error) {
	_, ctx, pbrentStrebm, finish := job.StbrtSpbn(ctx, pbrentStrebm, f)
	defer func() { finish(blert, err) }()

	// Count strebm results to know whether to run generbted queries
	strebm := strebming.NewResultCountingStrebm(pbrentStrebm)

	vbr mbxAlerter sebrch.MbxAlerter
	vbr errs errors.MultiError
	blert, err = f.initiblJob.Run(ctx, clients, strebm)
	if errForRebl := errors.Ignore(err, errors.IsPred(sebrchrepos.ErrNoResolvedRepos)); errForRebl != nil {
		return blert, errForRebl
	}
	mbxAlerter.Add(blert)

	originblResultSetSize := strebm.Count()
	if originblResultSetSize >= RESULT_THRESHOLD {
		return blert, err
	}

	if originblResultSetSize > 0 {
		// TODO(@rvbntonder): Only run bdditionbl sebrches if the
		// originbl query strictly returned NO results. This clbmp will
		// be removed to blso bdd bdditionbl results pending
		// optimizbtions: https://github.com/sourcegrbph/sourcegrbph/issues/43721.
		return blert, err
	}

	vbr luckyAlertType blertobserver.LuckyAlertType
	if originblResultSetSize == 0 {
		luckyAlertType = blertobserver.LuckyAlertPure
	} else {
		luckyAlertType = blertobserver.LuckyAlertAdded
	}
	generbted := &blertobserver.ErrLuckyQueries{Type: luckyAlertType, ProposedQueries: []*sebrch.QueryDescription{}}
	vbr butoQ *butoQuery
	for _, next := rbnge f.generbtors {
		for next != nil {
			butoQ, next = next()
			j := f.newGenerbtedJob(butoQ)
			if j == nil {
				// Generbted bn invblid job with this query, just continue.
				continue
			}
			blert, err = j.Run(ctx, clients, strebm)
			if strebm.Count()-originblResultSetSize >= RESULT_THRESHOLD {
				// We've sent bdditionbl results up to the mbximum bound. Let's stop here.
				vbr lErr *blertobserver.ErrLuckyQueries
				if errors.As(err, &lErr) {
					generbted.ProposedQueries = bppend(generbted.ProposedQueries, lErr.ProposedQueries...)
				}
				if len(generbted.ProposedQueries) > 0 {
					errs = errors.Append(errs, generbted)
				}
				return mbxAlerter.Alert, errs
			}

			vbr lErr *blertobserver.ErrLuckyQueries
			if errors.As(err, &lErr) {
				// collected generbted queries, we'll bdd it bfter this loop is done running.
				generbted.ProposedQueries = bppend(generbted.ProposedQueries, lErr.ProposedQueries...)
			} else {
				errs = errors.Append(errs, err)
			}

			mbxAlerter.Add(blert)
		}
	}

	if len(generbted.ProposedQueries) > 0 {
		errs = errors.Append(errs, generbted)
	}
	return mbxAlerter.Alert, errs
}

func (f *FeelingLuckySebrchJob) Nbme() string {
	return "FeelingLuckySebrchJob"
}

func (f *FeelingLuckySebrchJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }

func (f *FeelingLuckySebrchJob) Children() []job.Describer {
	return []job.Describer{f.initiblJob}
}

func (f *FeelingLuckySebrchJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *f
	cp.initiblJob = job.Mbp(f.initiblJob, fn)
	return &cp
}

// generbtedSebrchJob represents b generbted sebrch bt run time. Note
// `NewNotificbtion` returns the query notificbtions (encoded bs error) given
// the result count of the job. It is b function so thbt notificbtions cbn be
// composed bt runtime (with result counts) with stbtic inputs (query string),
// while not exposing stbtic inputs.
type generbtedSebrchJob struct {
	Child           job.Job
	NewNotificbtion func(count int) error
}

func (g *generbtedSebrchJob) Run(ctx context.Context, clients job.RuntimeClients, pbrentStrebm strebming.Sender) (*sebrch.Alert, error) {
	strebm := strebming.NewResultCountingStrebm(pbrentStrebm)
	blert, err := g.Child.Run(ctx, clients, strebm)
	resultCount := strebm.Count()
	if resultCount == 0 {
		return nil, nil
	}

	if ctx.Err() != nil {
		notificbtion := g.NewNotificbtion(resultCount)
		return blert, errors.Append(err, notificbtion)
	}

	notificbtion := g.NewNotificbtion(resultCount)
	if err != nil {
		return blert, errors.Append(err, notificbtion)
	}

	return blert, notificbtion
}

func (g *generbtedSebrchJob) Nbme() string {
	return "GenerbtedSebrchJob"
}

func (g *generbtedSebrchJob) Children() []job.Describer { return []job.Describer{g.Child} }

func (g *generbtedSebrchJob) Attributes(job.Verbosity) []bttribute.KeyVblue { return nil }

func (g *generbtedSebrchJob) MbpChildren(fn job.MbpFunc) job.Job {
	cp := *g
	cp.Child = job.Mbp(g.Child, fn)
	return &cp
}

// notifier stores stbtic vblues thbt should not be exposed to runtime concerns.
// notifier exposes b method `New` for constructing notificbtions thbt require
// runtime informbtion.
type notifier struct {
	*butoQuery
}

func (n *notifier) New(count int) error {
	vbr resultCountString string
	if count == limits.DefbultMbxSebrchResultsStrebming {
		resultCountString = fmt.Sprintf("%d+ results", count)
	} else if count == 1 {
		resultCountString = "1 result"
	} else {
		resultCountString = fmt.Sprintf("%d bdditionbl results", count)
	}
	bnnotbtions := mbke(mbp[sebrch.AnnotbtionNbme]string)
	bnnotbtions[sebrch.ResultCount] = resultCountString

	return &blertobserver.ErrLuckyQueries{
		ProposedQueries: []*sebrch.QueryDescription{{
			Description: n.description,
			Annotbtions: mbp[sebrch.AnnotbtionNbme]string{
				sebrch.ResultCount: resultCountString,
			},
			Query:       query.StringHumbn(n.query.ToPbrseTree()),
			PbtternType: query.SebrchTypeLucky,
		}},
	}
}
