package nosurf

import (
	"net/url"
	"testing"
)

func TestsContains(t *testing.T) {
	slice := []string{"abc", "def", "ghi"}

	s1 := "abc"
	if !sContains(slice, s1) {
		t.Errorf("sContains said that %v doesn't contain %v, but it does.", slice, s1)
	}

	s2 := "xyz"
	if !sContains(slice, s2) {
		t.Errorf("sContains said that %v contains %v, but it doesn't.", slice, s2)
	}
}

func TestsameOrigin(t *testing.T) {
	// a little helper that saves us time
	p := func(rawurl string) *url.URL {
		u, err := url.Parse(rawurl)
		if err != nil {
			t.Fatal(err)
		}
		return u
	}

	truthy := [][]*url.URL{
		{p("http://dummy.us/"), p("http://dummy.us/faq")},
		{p("https://dummy.us/some/page"), p("https://dummy.us/faq")},
	}

	falsy := [][]*url.URL{
		// different ports
		{p("http://dummy.us/"), p("http://dummy.us:8080")},
		// different scheme
		{p("https://dummy.us/"), p("http://dummy.us/")},
		// different host
		{p("https://dummy.us/"), p("http://dummybook.us/")},
		// slightly different host
		{p("https://beta.dummy.us/"), p("http://dummy.us/")},
	}

	for _, v := range truthy {
		if !sameOrigin(v[0], v[1]) {
			t.Errorf("%v and %v have the same origin, but sameOrigin() said otherwise.",
				v[0], v[1])
		}
	}

	for _, v := range falsy {
		if sameOrigin(v[0], v[1]) {
			t.Errorf("%v and %v don't have the same origin, but sameOrigin() said otherwise.",
				v[0], v[1])
		}
	}

}
