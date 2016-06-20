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

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/tools/godoc/vfs"
	"golang.org/x/tools/godoc/vfs/zipfs"

	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/errcode"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/handlerutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
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

	repo, repoRev, err := handlerutil.GetRepoAndRev(ctx, mux.Vars(r))
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
	importSizeBytes.WithLabelValues(repotrackutil.GetTrackedRepo(repo.URI)).Observe(float64(contentLength))

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
	fs := zipfs.New(&zip.ReadCloser{Reader: *zipR}, fmt.Sprintf("srclib import for %s", repo.URI))
	fs = absolutePathVFS{fs}

	// Import and index over gRPC.
	remoteStore := newSrclibStoreClient(ctx, pb.NewMultiRepoImporterClient(cl.Conn))

	importOpt := srclib.ImportOpt{
		Repo:     repo.URI,
		CommitID: repoRev.CommitID,
	}
	if err := srclib.Import(fs, remoteStore, importOpt); err != nil {
		return fmt.Errorf("srclib import of %s failed: %s", repo.URI, err)
	}

	// Update defs table in DB
	if _, err := cl.Defs.RefreshIndex(ctx, &sourcegraph.DefsRefreshIndexOp{
		Repo:     repoRev.Repo,
		CommitID: repoRev.CommitID,
	}); err != nil {
		return err
	}

	if repo.Fork {
		// Don't index forks in global search
		return nil
	}

	// global * reindex, doesn't block import
	cl.Async.RefreshIndexes(ctx, &sourcegraph.AsyncRefreshIndexesOp{Repo: repoRev.Repo, Force: true})

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

var importSizeBytes = prometheus.NewSummaryVec(prometheus.SummaryOpts{
	Namespace: "src",
	Subsystem: "srclib_import",
	Name:      "content_length_bytes",
	Help:      "Size of the request body (a zipfile) for srclib import in bytes",
}, []string{"repo"})

func init() {
	prometheus.MustRegister(importSizeBytes)
}
