//go:build go1.10
// +build go1.10

package search_test

import "archive/tar"

func addpaxheader(w *tar.Writer, body string) error {
	hdr := &tar.Header{
		Name:       "pax_global_header",
		Typeflag:   tar.TypeXGlobalHeader,
		PAXRecords: map[string]string{"somekey": body},
	}
	return w.WriteHeader(hdr)
}
