package httpapi

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	pathpkg "path"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/util/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/util/handlerutil"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
)

var newSrclibStoreClient = pb.Client // mockable for testing

// serveSrclibImport accepts a zip archive of a .srclib-cache
// directory and runs an import of the data into the repo rev
// specified in the URL route.
func serveSrclibImport(w http.ResponseWriter, r *http.Request) (err error) {
	// Check allowable content types and encodings.
	const allowedContentTypes = "|application/x-zip-compressed|application/x-zip|application/zip|application/octet-stream|"
	if ct := r.Header.Get("content-type"); !strings.Contains(allowedContentTypes, ct) || strings.Contains(ct, "|") {
		http.Error(w, "requires one of Content-Type: "+allowedContentTypes, http.StatusBadRequest)
		return nil
	}
	if strings.ToLower(r.Header.Get("content-transfer-encoding")) != "binary" {
		http.Error(w, "requires Content-Transfer-Encoding: binary", http.StatusBadRequest)
		return nil
	}

	ctx, cl := handlerutil.Client(r)

	_, repoRev, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
	if err != nil {
		return err
	}

	// Buffer the file to disk so we can provide zip.NewReader with a
	// io.ReaderAt.
	f, err := ioutil.TempFile("", "srclib-import")
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil && err == nil {
			err = err2
		}
		if err2 := os.Remove(f.Name()); err2 != nil && err == nil {
			err = err2
		}
	}()
	contentLength, err := io.Copy(f, r.Body)
	if err != nil {
		return err
	}
	if _, err := f.Seek(0, os.SEEK_SET); err != nil {
		return err
	}

	if contentLength == 0 {
		return &errcode.HTTPErr{Status: http.StatusBadRequest, Err: errors.New("no data in request body")}
	}

	zipR, err := zip.NewReader(f, contentLength)
	if err != nil {
		return err
	}

	// It's safe to construct the zip.ReadCloser without its private
	// *os.File field. If package zipfs's implementation changes in
	// such a way that makes this assumption false, our tests will
	// catch the issue.
	fs := zipfs.New(&zip.ReadCloser{Reader: *zipR}, fmt.Sprintf("srclib import for %s", repoRev))
	fs = absolutePathVFS{fs}

	// Import and index over gRPC.
	remoteStore := newSrclibStoreClient(ctx, pb.NewMultiRepoImporterClient(cl.Conn))

	importOpt := srclib.ImportOpt{
		Repo:     repoRev.URI,
		CommitID: repoRev.CommitID,
	}
	if err := srclib.Import(fs, remoteStore, importOpt); err != nil {
		return fmt.Errorf("srclib import of %s failed: %s", repoRev, err)
	}

	// Best-effort global search re-index, don't block import
	go func() {
		_, err := cl.Search.RefreshIndex(ctx, &sourcegraph.SearchRefreshIndexOp{
			Repos:         []*sourcegraph.RepoSpec{{repoRev.URI}},
			RefreshCounts: true,
			RefreshSearch: true,
		})
		if err != nil {
			log15.Error("search indexing failed", "repo", repoRev.URI, "commit", repoRev.CommitID, "err", err)
		}
	}()

	return nil
}

// absolutePathVFS translates relative paths to paths beginning with
// "/" (which the zipfs VFS requires).
type absolutePathVFS struct {
	vfs.FileSystem
}

func (fs absolutePathVFS) abs(path string) string {
	path = pathpkg.Clean(path)
	switch {
	case path == ".":
		return "/"
	case path[0] == '/':
		return path
	}
	return "/" + path
}

func (fs absolutePathVFS) Stat(path string) (os.FileInfo, error) {
	return fs.FileSystem.Stat(fs.abs(path))
}
func (fs absolutePathVFS) Lstat(path string) (os.FileInfo, error) {
	return fs.FileSystem.Lstat(fs.abs(path))
}
func (fs absolutePathVFS) Open(path string) (vfs.ReadSeekCloser, error) {
	return fs.FileSystem.Open(fs.abs(path))
}
func (fs absolutePathVFS) ReadDir(path string) ([]os.FileInfo, error) {
	return fs.FileSystem.ReadDir(fs.abs(path))
}
func (fs absolutePathVFS) String() string { return fmt.Sprintf("abs(%s)", fs.FileSystem) }
