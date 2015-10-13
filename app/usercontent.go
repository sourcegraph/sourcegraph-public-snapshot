package app

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/usercontent"
)

func serveUserContent(w http.ResponseWriter, req *http.Request) error {
	if usercontent.Store == nil {
		return fmt.Errorf("no store for user content available")
	}

	name := mux.Vars(req)["Name"]
	f, err := usercontent.Store.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	w.Header().Set("Content-Type", "image/png")
	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	return nil
}
