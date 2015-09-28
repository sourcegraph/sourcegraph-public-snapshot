package main

import (
	"net/http"

	"github.com/yosssi/ace"
)

// Pet represents a pet.
type Pet struct {
	Species string
	Name    string
	Age     int
}

func handler(w http.ResponseWriter, r *http.Request) {
	tpl, err := ace.Load("example", "", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Pets": []Pet{
			Pet{Species: "Dog", Name: "Taro", Age: 5},
			Pet{Species: "Cat", Name: "Hanako", Age: 10},
			Pet{Species: "Rabbit", Name: "Jiro", Age: 1},
		},
	}
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
