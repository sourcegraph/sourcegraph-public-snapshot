package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/satori/go.uuid"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/usercontent"
)

func serveUserContentUpload(w http.ResponseWriter, req *http.Request) error {
	const maxSizeBytes = 10 * 1024 * 1024

	if usercontent.Store == nil {
		return fmt.Errorf("no store for user content available")
	}

	if req.Header.Get("Content-Type") != "image/png" {
		return fmt.Errorf("invalid Content-Type: %v", w.Header().Get("Content-Type"))
	}
	body, err := ioutil.ReadAll(http.MaxBytesReader(w, req.Body, maxSizeBytes))
	if err != nil {
		return err
	}
	name := uuid.NewV4().String() + ".png"
	err = writeFile(usercontent.Store, name, body)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(struct {
		Name string
	}{
		Name: name,
	})
}

// writeFile writes data to a file named by name.
// If the file does not exist, writeFile creates it;
// otherwise writeFile truncates it before writing.
func writeFile(fs rwvfs.FileSystem, name string, data []byte) error {
	f, err := fs.Create(name)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
