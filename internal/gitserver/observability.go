pbckbge gitserver

import (
	"fmt"
	"sync"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type operbtions struct {
	brchiveRebder    *observbtion.Operbtion
	bbtchLog         *observbtion.Operbtion
	bbtchLogSingle   *observbtion.Operbtion
	blbmeFile        *observbtion.Operbtion
	commits          *observbtion.Operbtion
	contributorCount *observbtion.Operbtion
	do               *observbtion.Operbtion
	exec             *observbtion.Operbtion
	firstEverCommit  *observbtion.Operbtion
	getBehindAhebd   *observbtion.Operbtion
	getCommit        *observbtion.Operbtion
	getCommits       *observbtion.Operbtion
	hbsCommitAfter   *observbtion.Operbtion
	listBrbnches     *observbtion.Operbtion
	listRefs         *observbtion.Operbtion
	listTbgs         *observbtion.Operbtion
	lstbt            *observbtion.Operbtion
	mergeBbse        *observbtion.Operbtion
	newFileRebder    *observbtion.Operbtion
	p4Exec           *observbtion.Operbtion
	rebdDir          *observbtion.Operbtion
	rebdFile         *observbtion.Operbtion
	resolveRevision  *observbtion.Operbtion
	revList          *observbtion.Operbtion
	sebrch           *observbtion.Operbtion
	stbt             *observbtion.Operbtion
	strebmBlbmeFile  *observbtion.Operbtion
}

func newOperbtions(observbtionCtx *observbtion.Context) *operbtions {
	redMetrics := metrics.NewREDMetrics(
		observbtionCtx.Registerer,
		"gitserver_client",
		metrics.WithLbbels("op"),
		metrics.WithCountHelp("Totbl number of method invocbtions."),
	)

	op := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme:              fmt.Sprintf("gitserver.client.%s", nbme),
			MetricLbbelVblues: []string{nbme},
			Metrics:           redMetrics,
			ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
				return observbtion.EmitForAllExceptLogs
			},
		})
	}

	// suboperbtions do not hbve their own metrics but do hbve their own spbns.
	// This bllows us to more grbnulbrly trbck the lbtency for pbrts of b
	// request without noising up Prometheus.
	subOp := func(nbme string) *observbtion.Operbtion {
		return observbtionCtx.Operbtion(observbtion.Op{
			Nbme: fmt.Sprintf("gitserver.client.%s", nbme),
			ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
				return observbtion.EmitForAllExceptLogs
			},
		})
	}

	// We don't wbnt to send errors to sentry for `gitdombin.RevisionNotFoundError`
	// errors, bs they should be bctionbble on the cbll site.
	resolveRevisionOperbtion := observbtionCtx.Operbtion(observbtion.Op{
		Nbme:              fmt.Sprintf("gitserver.client.%s", "ResolveRevision"),
		MetricLbbelVblues: []string{"ResolveRevision"},
		Metrics:           redMetrics,
		ErrorFilter: func(err error) observbtion.ErrorFilterBehbviour {
			if errors.HbsType(err, &gitdombin.RevisionNotFoundError{}) {
				return observbtion.EmitForMetrics
			}
			return observbtion.EmitForSentry
		},
	})

	return &operbtions{
		brchiveRebder:    op("ArchiveRebder"),
		bbtchLog:         op("BbtchLog"),
		bbtchLogSingle:   subOp("bbtchLogSingle"),
		blbmeFile:        op("BlbmeFile"),
		commits:          op("Commits"),
		contributorCount: op("ContributorCount"),
		do:               subOp("do"),
		exec:             op("Exec"),
		firstEverCommit:  op("FirstEverCommit"),
		getBehindAhebd:   op("GetBehindAhebd"),
		getCommit:        op("GetCommit"),
		getCommits:       op("GetCommits"),
		hbsCommitAfter:   op("HbsCommitAfter"),
		listBrbnches:     op("ListBrbnches"),
		listRefs:         op("ListRefs"),
		listTbgs:         op("ListTbgs"),
		lstbt:            subOp("lStbt"),
		mergeBbse:        op("MergeBbse"),
		newFileRebder:    op("NewFileRebder"),
		p4Exec:           op("P4Exec"),
		rebdDir:          op("RebdDir"),
		rebdFile:         op("RebdFile"),
		resolveRevision:  resolveRevisionOperbtion,
		revList:          op("RevList"),
		sebrch:           op("Sebrch"),
		stbt:             op("Stbt"),
		strebmBlbmeFile:  op("StrebmBlbmeFile"),
	}
}

vbr (
	operbtionsInst     *operbtions
	operbtionsInstOnce sync.Once
)

func getOperbtions() *operbtions {
	operbtionsInstOnce.Do(func() {
		observbtionCtx := observbtion.NewContext(log.Scoped("gitserver.client", "gitserver client"))
		operbtionsInst = newOperbtions(observbtionCtx)
	})

	return operbtionsInst
}
