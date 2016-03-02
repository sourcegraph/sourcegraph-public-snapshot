package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/mux"

	"src.sourcegraph.com/sourcegraph/util"
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

var validateCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "src_srclib_validate",
	Help: "Srclib validate results, sanity checks the success/failure of srclib builds.",
},
	[]string{"status", "repo"},
)

func init() {
	prometheus.MustRegister(validateCounter)
}

func serveSrclibValidate(w http.ResponseWriter, r *http.Request) error {

	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
		http.Error(w, "requires Content-Type: application/json", http.StatusBadRequest)
		return nil
	}

	ctx, _ := handlerutil.Client(r)

	_, repoRev, _, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	dec := json.NewDecoder(r.Body)

	var val Validate
	err = dec.Decode(&val)
	if err != nil {
		return err
	}
	log15.Debug("srclib validate output", "repoRev", repoRev, "output", val)

	trackedRepo := util.GetTrackedRepo(r.URL.Path)

	var counter prometheus.Counter
	if len(val.Warnings) == 0 {
		counter, err = validateCounter.GetMetricWithLabelValues("success", trackedRepo)
	} else {
		counter, err = validateCounter.GetMetricWithLabelValues("failure", trackedRepo)
	}
	if err != nil {
		return err
	}
	counter.Inc()

	return nil
}
