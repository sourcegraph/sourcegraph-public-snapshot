package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"src.sourcegraph.com/sourcegraph/util/handlerutil"

	"gopkg.in/inconshreveable/log15.v2"
)

type Validate struct {
	Warnings []BuildWarning
}

type BuildWarning struct {
	Directory string
	Warning   string
}

var validateCounter *prometheus.CounterVec

func init() {
	validateCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "src_srclib_validate",
		Help: "Srclib validate results, sanity checks the success/failure of srclib builds.",
	},
		[]string{"repo_build_success"},
	)

	prometheus.MustRegister(validateCounter)
}

func serveSrclibCoverage(w http.ResponseWriter, r *http.Request) error {

	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
		http.Error(w, "requires Content-Type: application/json", http.StatusBadRequest)
		return nil
	}

	cl := handlerutil.APIClient(r)

	_, repoRev, _, err := handlerutil.GetRepoAndRev(r, cl.Repos)
	if err != nil {
		return err
	}

	dec := json.NewDecoder(r.Body)

	var val Validate
	err = dec.Decode(&val)
	if err != nil {
		return err
	}

	log15.Info("Srclib Validate Output", "repoRev", repoRev, "srclib validate output", val)

	var counter prometheus.Counter
	if len(val.Warnings) == 0 {
		counter, err = validateCounter.GetMetricWithLabelValues("success")
	} else {
		counter, err = validateCounter.GetMetricWithLabelValues("failure")
	}
	if err != nil {
		return err
	}
	counter.Inc()

	return nil
}
