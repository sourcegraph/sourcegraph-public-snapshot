package main

import (
	"net/http"

	"github.com/yosssi/ace"
)

func handler(w http.ResponseWriter, r *http.Request) {
	tpl, err := ace.Load("views/example", "", &ace.Options{
		Asset: Asset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
