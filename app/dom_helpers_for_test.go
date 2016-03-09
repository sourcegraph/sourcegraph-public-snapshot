package app_test

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(r *http.Response) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(r.Body)
}

func checkHeader(resp *http.Response, name, wantValue string) error {
	value := resp.Header.Get(name)
	if value == wantValue {
		return nil
	}
	return fmt.Errorf("got header %q = %q, want %q", name, value, wantValue)
}
