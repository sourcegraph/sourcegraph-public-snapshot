// Package gzipfileserver provides an http.Handler that serves the given
// virtual file system with gzip compression, without special handling
// of index.html, and without directory listing.
package gzipfileserver

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
)

type gzipFileServer struct {
	root http.FileSystem
}

// New returns a gzip file server, that serves the given virtual file system
// with gzip compression, without special handling of index.html, and without
// directory listing.
func New(root http.FileSystem) http.Handler {
	return &gzipFileServer{root: root}
}

func (fs *gzipFileServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, "/") {
		req.URL.Path = "/" + req.URL.Path
	}

	// name is '/'-separated, not filepath.Separator.
	name := path.Clean(req.URL.Path)

	f, err := fs.root.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	// redirect to canonical path: / at end of directory url
	// req.URL.Path always begins with /
	url := req.URL.Path
	if fi.IsDir() {
		if url[len(url)-1] != '/' {
			localRedirect(w, req, path.Base(url)+"/")
			return
		}
	} else {
		if url[len(url)-1] == '/' && url != "/" {
			localRedirect(w, req, "../"+path.Base(url))
			return
		}
	}

	// A directory?
	if fi.IsDir() {
		// gzipFileServer does not provide directory listings.
		http.Error(w, "404 Page Not Found", http.StatusNotFound)
		return
	}

	// If client doesn't accept gzip encoding, serve without compression.
	if !isGzipEncodingAccepted(req) {
		http.ServeContent(w, req, fi.Name(), fi.ModTime(), f)
		return
	}

	// If the file is not worth gzip compressing, serve it as is.
	type notWorthGzipCompressing interface {
		NotWorthGzipCompressing()
	}
	if _, ok := f.(notWorthGzipCompressing); ok {
		http.ServeContent(w, req, fi.Name(), fi.ModTime(), f)
		return
	}

	// If there are gzip encoded bytes available, use them directly.
	type gzipByter interface {
		GzipBytes() []byte
	}
	if gzipFile, ok := f.(gzipByter); ok {
		w.Header().Set("Content-Encoding", "gzip")
		http.ServeContent(w, req, fi.Name(), fi.ModTime(), bytes.NewReader(gzipFile.GzipBytes()))
		return
	}

	// Perform compression and serve gzip compressed bytes (if it's worth it).
	if rs, err := gzipCompress(f); err == nil {
		w.Header().Set("Content-Encoding", "gzip")
		http.ServeContent(w, req, fi.Name(), fi.ModTime(), rs)
		return
	}

	// Serve as is.
	http.ServeContent(w, req, fi.Name(), fi.ModTime(), f)
}

// gzipCompress compresses input from r and returns it as an io.ReadSeeker.
// It returns an error if compressed size is not smaller than uncompressed.
func gzipCompress(r io.Reader) (io.ReadSeeker, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	n, err := io.Copy(gw, r)
	if err != nil {
		return nil, err
	}
	err = gw.Close()
	if err != nil {
		return nil, err
	}
	if int64(buf.Len()) >= n {
		return nil, fmt.Errorf("not worth gzip compressing: original size %v, compressed size %v", n, buf.Len())
	}
	return bytes.NewReader(buf.Bytes()), nil
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

// toHTTPError returns a non-specific HTTP error message and status code
// for a given non-nil error value.
func toHTTPError(err error) (msg string, httpStatus int) {
	switch {
	case os.IsNotExist(err):
		return "404 Page Not Found", http.StatusNotFound
	case os.IsPermission(err):
		return "403 Forbidden", http.StatusForbidden
	default:
		return "500 Internal Server Error", http.StatusInternalServerError
	}
}

// isGzipEncodingAccepted returns true if the request includes "gzip" under Accept-Encoding header.
func isGzipEncodingAccepted(req *http.Request) bool {
	for _, v := range strings.Split(req.Header.Get("Accept-Encoding"), ",") {
		if strings.TrimSpace(v) == "gzip" {
			return true
		}
	}
	return false
}
