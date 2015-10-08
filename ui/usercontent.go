package ui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/docker/distribution/uuid"
	"github.com/shurcooL/webdavfs/vfsutil"
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
	name := uuid.Generate().String() + ".png"
	err = vfsutil.WriteFile(usercontent.Store, name, body, 0644)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(struct {
		Name string
	}{
		Name: name,
	})
}
