package app

import "net/http"

type coverage struct {
}

type repoCoverage struct {
}

func serveCoverage(w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("coverage placeholder"))
	return nil
}
