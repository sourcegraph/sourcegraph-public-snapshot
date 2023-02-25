package backend

import (
	"math/rand"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestConfigFingerprintChangesSince(t *testing.T) {
	mk := func(sc *schema.SiteConfiguration, timestamp time.Time) *ConfigFingerprint {
		t.Helper()
		fingerprint, err := NewConfigFingerprint(sc)
		if err != nil {
			t.Fatal(err)
		}

		fingerprint.ts = timestamp
		return fingerprint
	}

	zero := time.Time{}
	now := time.Now()
	oneDayAhead := now.Add(24 * time.Hour)

	fingerprintV1 := mk(&schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexBranches: map[string][]string{
				"foo": {"dev"},
			},
		},
	}, now)

	fingerprintV2 := mk(&schema.SiteConfiguration{
		ExperimentalFeatures: &schema.ExperimentalFeatures{
			SearchIndexBranches: map[string][]string{
				"foo": {"dev", "qa"},
			},
		},
	}, oneDayAhead)

	for _, tc := range []struct {
		name         string
		fingerPrintA *ConfigFingerprint
		fingerPrintB *ConfigFingerprint

		timeLowerBound time.Time
		timeUpperBound time.Time
	}{
		{
			name:         "missing fingerprint A",
			fingerPrintA: nil,
			fingerPrintB: fingerprintV1,

			timeLowerBound: zero,
			timeUpperBound: zero,
		},

		{
			name:         "missing fingerprint B",
			fingerPrintA: nil,
			fingerPrintB: fingerprintV1,

			timeLowerBound: zero,
			timeUpperBound: zero,
		},

		{
			name:         "same fingerprint",
			fingerPrintA: fingerprintV1,
			fingerPrintB: fingerprintV1,

			timeLowerBound: fingerprintV1.ts.Add(-3 * time.Minute),
			timeUpperBound: fingerprintV1.ts.Add(3 * time.Minute),
		},

		{
			name:         "different fingerprint",
			fingerPrintA: fingerprintV1,
			fingerPrintB: fingerprintV2,

			timeLowerBound: zero,
			timeUpperBound: zero,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.fingerPrintA.ChangesSince(tc.fingerPrintB)

			if tc.timeLowerBound.Equal(got) {
				return
			}

			if got.After(tc.timeLowerBound) && got.Before(tc.timeUpperBound) {
				return
			}

			t.Errorf("got %s, not in range [%s, %s)",
				got.Format(time.RFC3339),
				tc.timeLowerBound.Format(time.RFC3339),
				tc.timeUpperBound.Format(time.RFC3339),
			)
		})
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

	if !cfA.sameConfig(cfB) {
		t.Fatal("expected same config for A and B")
	}
	if cfA.sameConfig(cfC) {
		t.Fatal("expected different config for A and C")
	}
}

func TestSiteConfigFingerprint_RoundTrip(t *testing.T) {
	type roundTripper func(t *testing.T, original *ConfigFingerprint) (converted *ConfigFingerprint)

	for _, test := range []struct {
		transportName string
		roundTripper  roundTripper
	}{
		{
			transportName: "gRPC",
			roundTripper: func(_ *testing.T, original *ConfigFingerprint) *ConfigFingerprint {
				converted := &ConfigFingerprint{}
				converted.FromProto(original.ToProto())
				return converted
			},
		},
		{

			transportName: "HTTP headers",
			roundTripper: func(t *testing.T, original *ConfigFingerprint) *ConfigFingerprint {
				echoHandler := func(w http.ResponseWriter, r *http.Request) {
					// echo back the fingerprint in the response
					var fingerprint ConfigFingerprint
					err := fingerprint.FromHeaders(r.Header)
					if err != nil {
						t.Fatalf("error converting from request in echoHandler: %s", err)
					}

					fingerprint.ToHeaders(w.Header())
				}

				w := httptest.NewRecorder()

				r := httptest.NewRequest("GET", "/", nil)
				original.ToHeaders(r.Header)
				echoHandler(w, r)

				var converted ConfigFingerprint
				err := converted.FromHeaders(w.Result().Header)
				if err != nil {
					t.Fatalf("error converting from response outside echoHandler: %s", err)
				}

				return &converted
			},
		},
	} {
		t.Run(test.transportName, func(t *testing.T) {
			var diff string
			f := func(ts fuzzTime, hash uint64) bool {
				original := ConfigFingerprint{
					ts:   time.Time(ts),
					hash: hash,
				}

				converted := test.roundTripper(t, &original)

				if diff = cmp.Diff(original.hash, converted.hash); diff != "" {
					diff = "hash: " + diff
					return false
				}

				if diff = cmp.Diff(original.ts, converted.ts, cmpopts.EquateApproxTime(time.Second)); diff != "" {
					diff = "ts: " + diff
					return false
				}

				return true
			}

			if err := quick.Check(f, nil); err != nil {
				t.Errorf("fingerprint diff (-want +got):\n%s", diff)
			}
		})
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

func testMarshal(t *testing.T, cf *ConfigFingerprint) {
	t.Helper()

	v := cf.Marshal()
	t.Log(v)

	got, err := parseConfigFingerprint(v)
	if err != nil {
		t.Fatal(err)
	}

	if !cf.sameConfig(got) {
		t.Error("expected same config")
	}

	since := got.paddedTimestamp()
	if since.After(cf.ts) {
		t.Error("since should not be after Timestamp")
	}
	if since.Before(cf.ts.Add(-time.Hour)) {
		t.Error("since should not be before Timestamp - hour")
	}
}

type fuzzTime time.Time

func (fuzzTime) Generate(rand *rand.Rand, _ int) reflect.Value {
	// The maximum representable year in RFC 3339 is 9999, so we'll use that as our upper bound.
	maxDate := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

	ts := time.Unix(rand.Int63n(maxDate.Unix()), rand.Int63n(int64(time.Second)))
	return reflect.ValueOf(fuzzTime(ts))
}

var _ quick.Generator = fuzzTime{}
