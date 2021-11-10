package backend

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestParseAndSetConfigFingerprint(t *testing.T) {
	mk := func(sc *schema.SiteConfiguration) *configFingerprint {
		t.Helper()
		fingerprint, err := newConfigFingerprint(sc)
		if err != nil {
			t.Fatal(err)
		}
		return fingerprint
	}

	parseAndSet := func(fingerprint string, sc *schema.SiteConfiguration) time.Time {
		t.Helper()
		r := httptest.NewRequest("GET", "/", nil)
		r.Header.Set("X-Sourcegraph-Config-Fingerprint", fingerprint)
		w := httptest.NewRecorder()
		minLastChanged, err := ParseAndSetConfigFingerprint(w, r, sc)
		if err != nil {
			t.Fatal(err)
		}

		got, err := parseConfigFingerprint(w.Result().Header.Get("X-Sourcegraph-Config-Fingerprint"))
		if err != nil {
			t.Fatal(err)
		}
		want := mk(sc)
		if !got.SameConfig(want) {
			t.Fatal("expected same config in response fingerprint")
		}

		return minLastChanged
	}

	sc1 := &schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexBranches: map[string][]string{
				"foo": {"dev"},
			},
		},
	}
	sc2 := &schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexBranches: map[string][]string{
				"foo": {"dev", "qa"},
			},
		},
	}

	if got := parseAndSet("", sc1); !got.IsZero() {
		t.Fatal("expect no min last changed for missing fingerprint")
	}

	if got := parseAndSet(mk(sc1).Marshal(), sc2); !got.IsZero() {
		t.Fatal("expect no min last changed for different site config")
	}

	if got := parseAndSet(mk(sc1).Marshal(), sc1); got.IsZero() {
		t.Fatal("expect min last changed for same site config")
	}
}

func TestConfigFingerprint(t *testing.T) {
	sc1 := &schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexBranches: map[string][]string{
				"foo": {"dev"},
			},
		},
	}
	sc2 := &schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexBranches: map[string][]string{
				"foo": {"dev", "qa"},
			},
		},
	}

	var seq time.Duration
	mk := func(sc *schema.SiteConfiguration) *configFingerprint {
		t.Helper()
		cf, err := newConfigFingerprint(sc)
		if err != nil {
			t.Fatal(err)
		}

		// Each consecutive call we adjust the time significantly to ensure
		// when comparing config we don't take into account time.
		cf.ts = cf.ts.Add(seq * time.Hour)
		seq++

		testMarshal(t, cf)
		return cf
	}

	cfA := mk(sc1)
	cfB := mk(sc1)
	cfC := mk(sc2)

	if !cfA.SameConfig(cfB) {
		t.Fatal("expected same config for A and B")
	}
	if cfA.SameConfig(cfC) {
		t.Fatal("expected different config for A and C")
	}
}

func TestConfigFingerprint_Marshal(t *testing.T) {
	// Use a fixed time for this test case
	now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if err != nil {
		t.Fatal(err)
	}

	cf := configFingerprint{
		ts:   now,
		hash: 123,
	}

	got := cf.Marshal()
	want := "search-config-fingerprint 1 2006-01-02T15:04:05Z 7b"
	if got != want {
		t.Errorf("unexpected marshal value:\ngot:  %s\nwant: %s", got, want)
	}

	testMarshal(t, &cf)
}

func TestConfigFingerprint_parse(t *testing.T) {
	ignore := []string{
		// we ignore empty / not specified
		"",
		// we ignore different versions
		"search-config-fingerprint 0",
		"search-config-fingerprint 2",
		"search-config-fingerprint 0 2006-01-02T15:04:05Z 7b",
		"search-config-fingerprint 2 2006-01-02T15:04:05Z 7b",
	}
	valid := []string{
		"search-config-fingerprint 1 2006-01-02T15:04:05Z 7b",
	}
	malformed := []string{
		"foobar",
		"search-config-fingerprint 1 2006-01-02T15:04:05Z",
		"search-config 1 2006-01-02T15:04:05Z 7b",
		"search-config 1 1 2",
	}

	for _, v := range ignore {
		cf, err := parseConfigFingerprint(v)
		if err != nil {
			t.Fatalf("unexpected error parsing ignorable %q: %v", v, err)
		}
		if !cf.ts.IsZero() {
			t.Fatalf("expected ignorable %q", v)
		}
	}
	for _, v := range valid {
		cf, err := parseConfigFingerprint(v)
		if err != nil {
			t.Fatalf("unexpected error parsing valid %q: %v", v, err)
		}
		if cf.ts.IsZero() {
			t.Fatalf("expected valid %q", v)
		}
	}
	for _, v := range malformed {
		_, err := parseConfigFingerprint(v)
		if err == nil {
			t.Fatalf("expected malformed %q", v)
		}
	}
}

func testMarshal(t *testing.T, cf *configFingerprint) {
	t.Helper()

	v := cf.Marshal()
	t.Log(v)

	got, err := parseConfigFingerprint(v)
	if err != nil {
		t.Fatal(err)
	}

	if !cf.SameConfig(got) {
		t.Error("expected same config")
	}

	since := got.Since()
	if since.After(cf.ts) {
		t.Error("since should not be after ts")
	}
	if since.Before(cf.ts.Add(-time.Hour)) {
		t.Error("since should not be before ts - hour")
	}
}
