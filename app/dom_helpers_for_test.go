package app_test

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseHTML(r *http.Response) (*goquery.Document, error) {
	return goquery.NewDocumentFromReader(r.Body)
}

func checkPageTitle(resp *http.Response, wantTitleContaining string) error {
	doc, err := parseHTML(resp)
	if err != nil {
		return err
	}
	if title := doc.Find("title").Text(); !strings.Contains(title, wantTitleContaining) {
		return fmt.Errorf("got page title %q, want it to contain %q", title, wantTitleContaining)
	}
	return nil
}

func checkHeader(resp *http.Response, name, wantValue string) error {
	value := resp.Header.Get(name)
	if value == wantValue {
		return nil
	}
	return fmt.Errorf("got header %q = %q, want %q", name, value, wantValue)
}
