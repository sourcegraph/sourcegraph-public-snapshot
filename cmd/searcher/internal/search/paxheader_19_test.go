//go:build !go1.10
// +build !go1.10

package search_test

import "archive/tar"

func addpaxheader(w *tar.Writer, body string) error {
	hdr := &tar.Header{
		Name:     "pax_global_header",
		Size:     int64(len(body)),
		Typeflag: tar.TypeXGlobalHeader,
	}
	if err := w.WriteHeader(hdr); err != nil {
		return err
	}
	_, err := w.Write([]byte(body))
	return err
}
