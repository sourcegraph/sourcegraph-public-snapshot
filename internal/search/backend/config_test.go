package backend

import (
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/schema"
)

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
	mk := func(sc *schema.SiteConfiguration) *ConfigFingerprint {
		t.Helper()
		cf, err := NewConfigFingerprint(sc)
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

	cf := ConfigFingerprint{
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

func testMarshal(t *testing.T, cf *ConfigFingerprint) {
	t.Helper()

	v := cf.Marshal()
	t.Log(v)

	var got ConfigFingerprint
	if err := got.Unmarshal(v); err != nil {
		t.Fatal(err)
	}

	if !cf.SameConfig(&got) {
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
