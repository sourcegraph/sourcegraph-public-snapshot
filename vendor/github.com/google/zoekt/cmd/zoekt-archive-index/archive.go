package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Archive interface {
	Next() (*File, error)
	Close() error
}

type File struct {
	io.Reader
	Name string
	Size int64
}

type tarArchive struct {
	tr     *tar.Reader
	closer io.Closer
}

func (a *tarArchive) Next() (*File, error) {
	for {
		hdr, err := a.tr.Next()
		if err != nil {
			return nil, err
		}

		// We only care about files
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}

		return &File{
			Reader: a.tr,
			Name:   hdr.Name,
			Size:   hdr.Size,
		}, nil
	}
}

func (a *tarArchive) Close() error {
	if a.closer != nil {
		return a.closer.Close()
	}
	return nil
}

// openArchive opens the tar at the URL or filepath u. Also supported is tgz
// files over http.
func openArchive(u string) (Archive, error) {
	var (
		r      io.Reader
		closer io.Closer
	)

	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		resp, err := http.Get(u)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			b, err := ioutil.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			if err != nil {
				return nil, err
			}
			return nil, &url.Error{
				Op:  "Get",
				URL: u,
				Err: fmt.Errorf("%s: %s", resp.Status, string(b)),
			}
		}
		closer = resp.Body
		r = resp.Body
		if resp.Header.Get("Content-Type") == "application/x-gzip" {
			r, err = gzip.NewReader(r)
			if err != nil {
				resp.Body.Close()
				return nil, err
			}
		}
	} else if u == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(u)
		if err != nil {
			return nil, err
		}
		closer = f
		r = f
	}

	return &tarArchive{
		tr:     tar.NewReader(r),
		closer: closer,
	}, nil
}
