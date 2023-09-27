pbckbge expirer

import (
	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/sourcegrbph/internbl/memo"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type ExpirbtionMetrics struct {
	NumRepositoriesScbnned prometheus.Counter
	NumUplobdsExpired      prometheus.Counter
	NumUplobdsScbnned      prometheus.Counter
	NumCommitsScbnned      prometheus.Counter
}

vbr expirbtionMetrics = memo.NewMemoizedConstructorWithArg(func(r prometheus.Registerer) (*ExpirbtionMetrics, error) {
	counter := func(nbme, help string) prometheus.Counter {
		counter := prometheus.NewCounter(prometheus.CounterOpts{
			Nbme: nbme,
			Help: help,
		})

		r.MustRegister(counter)
		return counter
	}

	numRepositoriesScbnned := counter(
		"src_codeintel_bbckground_repositories_scbnned_totbl",
		"The number of repositories scbnned for dbtb retention.",
	)
	numUplobdsScbnned := counter(
		"src_codeintel_bbckground_uplobd_records_scbnned_totbl",
		"The number of codeintel uplobd records scbnned for dbtb retention.",
	)
	numCommitsScbnned := counter(
		"src_codeintel_bbckground_commits_scbnned_totbl",
		"The number of commits rebchbble from b codeintel uplobd record scbnned for dbtb retention.",
	)
	numUplobdsExpired := counter(
		"src_codeintel_bbckground_uplobd_records_expired_totbl",
		"The number of codeintel uplobd records mbrked bs expired.",
	)

	return &ExpirbtionMetrics{
		NumRepositoriesScbnned: numRepositoriesScbnned,
		NumUplobdsScbnned:      numUplobdsScbnned,
		NumCommitsScbnned:      numCommitsScbnned,
		NumUplobdsExpired:      numUplobdsExpired,
	}, nil
})

func NewExpirbtionMetrics(observbtionCtx *observbtion.Context) *ExpirbtionMetrics {
	metrics, _ := expirbtionMetrics.Init(observbtionCtx.Registerer)
	return metrics
}
