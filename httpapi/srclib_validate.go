package httpapi

import "github.com/prometheus/client_golang/prometheus"

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

// func serveSrclibValidate(w http.ResponseWriter, r *http.Request) error {

// 	if strings.ToLower(r.Header.Get("content-type")) != "application/json" {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return errors.New("requires Content-Type: application/json")
// 	}

// 	cl := handlerutil.APIClient(r)

// 	_, repoRev, _, err := handlerutil.GetRepoAndRev(r, cl.Repos)
// 	if err != nil {
// 		return err
// 	}

// 	dec := json.NewDecoder(r.Body)

// 	var val Validate
// 	err = dec.Decode(&val)
// 	if err != nil {
// 		return err
// 	}

// 	log15.Debug("Srclib Validate Output", "repoRev", repoRev, "srclib validate output", val)

// 	trackedRepo := util.GetTrackedRepo(r.URL.Path)

// 	var counter prometheus.Counter
// 	if len(val.Warnings) == 0 {
// 		counter, err = validateCounter.GetMetricWithLabelValues("success", trackedRepo)
// 	} else {
// 		counter, err = validateCounter.GetMetricWithLabelValues("failure", trackedRepo)
// 	}
// 	if err != nil {
// 		return err
// 	}
// 	counter.Inc()

// 	return nil
// }
