package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/grafana/regexp"

	"github.com/google/go-cmp/cmp"
)

func TestRender(t *testing.T) {
	full := &searchResultsAlert{
		Title:       "foo",
		Description: "bar",
		ProposedQueries: []ProposedQuery{
			{
				Description: "quux",
				Query:       "xyz:abc",
			},
			{
				Description: "baz",
				Query:       "def:ghi",
			},
		},
	}

	zero := &searchResultsAlert{}

	t.Run("zero value", func(t *testing.T) {
		content, err := zero.Render()
		if err != nil {
			t.Errorf("unexpected error rendering zero alert: %v", err)
		}

		if content != "" {
			t.Errorf("unexpected content for zero alert: %s", content)
		}
	})

	t.Run("no description", func(t *testing.T) {
		alert := *full
		alert.Description = zero.Description

		content, err := alert.Render()
		if err != nil {
			t.Errorf("unexpected error rendering alert without description: %v", err)
		}

		expected := subColorCodes(strings.Join([]string{
			"[[search-alert-title]]❗foo[[nc]]\n",
			"[[search-alert-proposed-title]]  Did you mean:[[nc]]\n",
			"[[search-alert-proposed-query]]  xyz:abc[[nc]] - [[search-alert-proposed-description]]quux[[nc]]\n",
			"[[search-alert-proposed-query]]  def:ghi[[nc]] - [[search-alert-proposed-description]]baz[[nc]]\n",
		}, ""))
		if diff := cmp.Diff(expected, content); diff != "" {
			t.Errorf("alert without description content incorrect: %s", diff)
		}
	})

	t.Run("no proposed queries", func(t *testing.T) {
		alert := *full
		alert.ProposedQueries = zero.ProposedQueries

		content, err := alert.Render()
		if err != nil {
			t.Errorf("unexpected error rendering alert without queries: %v", err)
		}

		expected := subColorCodes(strings.Join([]string{
			"[[search-alert-title]]❗foo[[nc]]\n",
			"[[search-alert-description]]  bar[[nc]]\n",
		}, ""))
		if diff := cmp.Diff(expected, content); diff != "" {
			t.Errorf("alert without queries content incorrect: %s", diff)
		}
	})

	t.Run("full", func(t *testing.T) {
		content, err := full.Render()
		if err != nil {
			t.Errorf("unexpected error rendering full alert: %v", err)
		}

		expected := subColorCodes(strings.Join([]string{
			"[[search-alert-title]]❗foo[[nc]]\n",
			"[[search-alert-description]]  bar[[nc]]\n",
			"[[search-alert-proposed-title]]  Did you mean:[[nc]]\n",
			"[[search-alert-proposed-query]]  xyz:abc[[nc]] - [[search-alert-proposed-description]]quux[[nc]]\n",
			"[[search-alert-proposed-query]]  def:ghi[[nc]] - [[search-alert-proposed-description]]baz[[nc]]\n",
		}, ""))
		if diff := cmp.Diff(expected, content); diff != "" {
			t.Errorf("full alert content incorrect: %s", diff)
		}
	})
}

var subColorCodesRegex = regexp.MustCompile(`\[\[[a-zA-Z0-9-]+\]\]`)

// subColorCodes provides ad-hoc templating of just ANSI colour codes from our
// ansiColors map, using a [[colour-code]] syntax.
func subColorCodes(in string) string {
	// We could use text/template here, but at a certain point it feels like
	// we're just reinventing the template string that's a const in
	// search_alert.go. This allows us to express the colour codes in the
	// expected string literals above while hopefully maintaining some meaning
	// to the tests.
	return subColorCodesRegex.ReplaceAllStringFunc(in, func(code string) string {
		esc, ok := ansiColors[strings.Trim(code, "[]")]
		if !ok {
			panic(fmt.Sprintf("cannot find colour %s", code))
		}
		return esc
	})
}
