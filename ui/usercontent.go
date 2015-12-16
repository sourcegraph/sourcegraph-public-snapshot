package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/satori/go.uuid"
	"sourcegraph.com/sourcegraph/rwvfs"
	"src.sourcegraph.com/sourcegraph/auth"
	"src.sourcegraph.com/sourcegraph/usercontent"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

var allowedMimeTypes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/gif":  {},
}

func serveUserContentUpload(w http.ResponseWriter, req *http.Request) error {
	const maxSizeBytes = 10 * 1024 * 1024

	// TODO: We should be doing gRPC calls for storing content, and keep
	//       the webserver stateless.

	actor := auth.ActorFromContext(httpctx.FromRequest(req))
	if !actor.HasWriteAccess() {
		return fmt.Errorf("actor (uid=%v, domain=%v) doesn't have write access", actor.UID, actor.Domain)
	}

	if usercontent.Store == nil {
		return fmt.Errorf("no store for user content available")
	}

	body, err := ioutil.ReadAll(http.MaxBytesReader(w, req.Body, maxSizeBytes))
	if err != nil {
		return err
	}

	mimeType := http.DetectContentType(body)
	_, ok := allowedMimeTypes[mimeType]
	if !ok {
		return fmt.Errorf("unsupported mime type: %v", mimeType)
	}

	extensions, err := mime.ExtensionsByType(mimeType)
	if err != nil {
		return err
	}

	extension := ""
	if extensions != nil {
		extension = extensions[0]
	} else {
		return fmt.Errorf("unable to calculate extension")
	}

	name := uuid.NewV4().String() + extension
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
