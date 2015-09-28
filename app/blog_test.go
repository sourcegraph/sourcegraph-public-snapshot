// These tests hit the live Tumblr API using the network, so they're
// slow.
//
// +build nettest

package app_test

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/app/internal/appconf"
	"src.sourcegraph.com/sourcegraph/app/internal/apptest"
	"src.sourcegraph.com/sourcegraph/app/router"
)

// Use real data because this is currently not easy to mock.
var sampleBlogPostSlug = "most-popular-django-model-field"

func TestBlogIndex(t *testing.T) {
	appconf.Current.Blog = true
	c, _ := apptest.New()

	resp, err := c.Get(router.Rel.URLTo(router.BlogIndex).String())
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPageTitle(resp, "Blog"); err != nil {
		t.Error(err)
	}
	if err := checkHeader(resp, "content-type", "text/html; charset=utf-8"); err != nil {
		t.Error(err)
	}
}

func TestBlogIndex_Atom(t *testing.T) {
	appconf.Current.Blog = true
	c, _ := apptest.New()

	resp, err := c.Get(router.Rel.URLToBlogAtomFeed().String())
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPageTitle(resp, "Blog"); err != nil {
		t.Error(err)
	}
	if err := checkHeader(resp, "content-type", "application/atom+xml; charset=utf-8"); err != nil {
		t.Error(err)
	}
}

func TestBlogPost(t *testing.T) {
	appconf.Current.Blog = true
	c, _ := apptest.New()

	resp, err := c.Get(router.Rel.URLToBlogPost(sampleBlogPostSlug).String())
	if err != nil {
		t.Fatal(err)
	}
	if err := checkPageTitle(resp, "Blog"); err != nil {
		t.Error(err)
	}
	if err := checkHeader(resp, "content-type", "text/html; charset=utf-8"); err != nil {
		t.Error(err)
	}
}
