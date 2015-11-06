// Package gzip_file_server provides a http.Handler that serves the given virtual file system with gzip compression,
// without special handling of index.html, and detailed HTTP error messages.
package gzip_file_server

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type gzipFileServer struct {
	root http.FileSystem
}

// New returns a gzip file server, that serves the given virtual file system without special handling of index.html.
func New(root http.FileSystem) http.Handler {
	return &gzipFileServer{root: root}
}

func (f *gzipFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/") {
		r.URL.Path = "/" + r.URL.Path
	}
	f.serveFile(w, r, path.Clean(r.URL.Path))
}

// name is '/'-separated, not filepath.Separator.
func (fs *gzipFileServer) serveFile(w http.ResponseWriter, req *http.Request, name string) {
	f, err := fs.root.Open(name)
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}
	defer f.Close()

	d, err := f.Stat()
	if err != nil {
		msg, code := toHTTPError(err)
		http.Error(w, msg, code)
		return
	}

	// redirect to canonical path: / at end of directory url
	// req.URL.Path always begins with /
	url := req.URL.Path
	if d.IsDir() {
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
	if d.IsDir() {
		if checkLastModified(w, req, d.ModTime()) {
			return
		}
		dirList(w, f, name)
		return
	}

	/*if _, plain := req.URL.Query()["plain"]; plain {
		w.Header().Set("Content-Type", "text/plain")
	}*/

	// If client doesn't accept gzip encoding, serve without compression.
	if !isGzipEncodingAccepted(req) {
		http.ServeContent(w, req, d.Name(), d.ModTime(), f)
		return
	}

	// If the file is not worth gzip compressing, serve it as is.
	type notWorthGzipCompressing interface {
		NotWorthGzipCompressing()
	}
	if _, ok := f.(notWorthGzipCompressing); ok {
		http.ServeContent(w, req, d.Name(), d.ModTime(), f)
		return
	}

	// If there are gzip encoded bytes available, use them directly.
	type gzipByter interface {
		GzipBytes() []byte
	}
	if gzipFile, ok := f.(gzipByter); ok {
		w.Header().Set("Content-Encoding", "gzip")
		http.ServeContent(w, req, d.Name(), d.ModTime(), bytes.NewReader(gzipFile.GzipBytes()))
		return
	}

	// Perform compression and serve gzip compressed bytes (if it's worth it).
	if rs, err := gzipCompress(f); err == nil {
		w.Header().Set("Content-Encoding", "gzip")
		http.ServeContent(w, req, d.Name(), d.ModTime(), rs)
		return
	}

	// Serve as is.
	http.ServeContent(w, req, d.Name(), d.ModTime(), f)
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

func dirList(w http.ResponseWriter, f http.File, name string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	switch name {
	case "/":
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", "/", ".")
	default:
		fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", path.Clean(name+"/.."), "..")
	}
	for {
		dirs, err := f.Readdir(100)
		if err != nil || len(dirs) == 0 {
			break
		}
		sort.Sort(byName(dirs))
		for _, d := range dirs {
			name := d.Name()
			if d.IsDir() {
				name += "/"
			}
			// name may contain '?' or '#', which must be escaped to remain
			// part of the URL path, and not indicate the start of a query
			// string or fragment.
			url := url.URL{Path: name}
			fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", url.String(), html.EscapeString(name))
		}
	}
	fmt.Fprintf(w, "</pre>\n")
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

// modtime is the modification time of the resource to be served, or IsZero().
// return value is whether this request is now complete.
func checkLastModified(w http.ResponseWriter, r *http.Request, modtime time.Time) bool {
	if modtime.IsZero() {
		return false
	}

	// The Date-Modified header truncates sub-second precision, so
	// use mtime < t+1s instead of mtime <= t to check for unmodified.
	if t, err := time.Parse(http.TimeFormat, r.Header.Get("If-Modified-Since")); err == nil && modtime.Before(t.Add(1*time.Second)) {
		h := w.Header()
		delete(h, "Content-Type")
		delete(h, "Content-Length")
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("Last-Modified", modtime.UTC().Format(http.TimeFormat))
	return false
}

// byName implements sort.Interface.
type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// toHTTPError returns a detailed HTTP error message and status code
// for a given non-nil error value.
func toHTTPError(err error) (msg string, httpStatus int) {
	switch {
	case os.IsNotExist(err):
		return fmt.Sprintf("404 Page Not Found\n\n%v", err), http.StatusNotFound
	case os.IsPermission(err):
		return fmt.Sprintf("403 Forbidden\n\n%v", err), http.StatusForbidden
	default:
		return fmt.Sprintf("500 Internal Server Error\n\n%v", err), http.StatusInternalServerError
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
