package main

import "net/http"

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
