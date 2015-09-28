package main

import (
	"html/template"
	"net/http"

	"github.com/yosssi/ace"
)

func handler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"Greeting": func(s string) string {
			return "Hello " + s
		},
	}
	tpl, err := ace.Load("example", "", &ace.Options{
		FuncMap: funcMap,
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
