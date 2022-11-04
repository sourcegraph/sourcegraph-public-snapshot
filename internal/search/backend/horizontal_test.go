package backend

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/zoekt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestHorizontalSearcher(t *testing.T) {
	var endpoints atomicMap
	endpoints.Store(prefixMap{})

	searcher := &HorizontalSearcher{
		Map: &endpoints,
		Dial: func(endpoint string) zoekt.Streamer {
			repoID, _ := strconv.Atoi(endpoint)
			var rle zoekt.RepoListEntry
			rle.Repository.Name = endpoint
			rle.Repository.ID = uint32(repoID)
			client := &FakeSearcher{
				Result: &zoekt.SearchResult{
					Files: []zoekt.FileMatch{{
						Repository: endpoint,
					}},
				},
				Repos: []*zoekt.RepoListEntry{&rle},
			}
			// Return metered searcher to test that codepath
			return NewMeteredSearcher(endpoint, client)
		},
	}
	defer searcher.Close()

	// Start up background goroutines which continuously hit the searcher
	// methods to ensure we are safe under concurrency.
	for i := 0; i < 5; i++ {
		cleanup := backgroundSearch(searcher)
		defer cleanup(t)
	}

	// each map is the set of servers at a point in time. This is to mainly
	// stress the management code.
	maps := []prefixMap{
		// Start with a normal config of two replicas
		{"1", "2"},

		// Add two
		{"1", "2", "3", "4"},

		// Lose two
		{"2", "4"},

		// Lose and add
		{"1", "2"},

		// Lose all
		{},

		// Lots
		{"1", "2", "3", "4", "5", "6", "7", "8", "9"},
	}

	for _, m := range maps {
		t.Log("current", searcher.String(), "next", m)
		endpoints.Store(m)

		// Our search results should be one per server
		sr, err := searcher.Search(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, fm := range sr.Files {
			got = append(got, fm.Repository)
		}
		sort.Strings(got)
		want := []string(m)
		if !cmp.Equal(want, got, cmpopts.EquateEmpty()) {
			t.Errorf("search mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}

		// Our list results should be one per server
		rle, err := searcher.List(context.Background(), nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		got = []string{}
		for _, r := range rle.Repos {
			got = append(got, r.Repository.Name)
		}
		sort.Strings(got)
		if !cmp.Equal(want, got, cmpopts.EquateEmpty()) {
			t.Errorf("list mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}

		rle, err = searcher.List(context.Background(), nil, &zoekt.ListOptions{Minimal: true})
		if err != nil {
			t.Fatal(err)
		}
		got = []string{}
		for r := range rle.Minimal {
			got = append(got, strconv.Itoa(int(r)))
		}
		sort.Strings(got)
		if !cmp.Equal(want, got, cmpopts.EquateEmpty()) {
			t.Fatalf("list mismatch (-want +got):\n%s", cmp.Diff(want, got))
		}
	}

	searcher.Close()
}

func TestDoStreamSearch(t *testing.T) {
	var endpoints atomicMap
	endpoints.Store(prefixMap{"1"})

	searcher := &HorizontalSearcher{
		Map: &endpoints,
		Dial: func(endpoint string) zoekt.Streamer {
			client := &FakeSearcher{
				SearchError: errors.Errorf("test error"),
			}
			// Return metered searcher to test that codepath
			return NewMeteredSearcher(endpoint, client)
		},
	}
	defer searcher.Close()

	c := make(chan *zoekt.SearchResult)
	defer close(c)
	err := searcher.StreamSearch(
		context.Background(),
		nil,
		nil,
		ZoektStreamFunc(func(event *zoekt.SearchResult) { c <- event }),
	)
	if err == nil {
		t.Fatalf("received non-nil error, but expected an error")
	}
}

func TestSyncSearchers(t *testing.T) {
	// This test exists to ensure we test the slow path for
	// HorizontalSearcher.searchers. The slow-path is
	// syncSearchers. TestHorizontalSearcher tests the same code paths, but
	// isn't guaranteed to trigger the all the parts of syncSearchers.
	var endpoints atomicMap
	endpoints.Store(prefixMap{"a"})

	type mock struct {
		FakeSearcher
		dialNum int
	}

	dialNumCounter := 0
	searcher := &HorizontalSearcher{
		Map: &endpoints,
		Dial: func(endpoint string) zoekt.Streamer {
			dialNumCounter++
			return &mock{
				dialNum: dialNumCounter,
			}
		},
	}
	defer searcher.Close()

	// First call initializes the list, second should use the fast-path so
	// should have the same dialNum.
	for i := 0; i < 2; i++ {
		t.Log("gen", i)
		m, err := searcher.syncSearchers()
		if err != nil {
			t.Fatal(err)
		}
		if len(m) != 1 {
			t.Fatal(err)
		}
		if got, want := m["a"].(*mock).dialNum, 1; got != want {
			t.Fatalf("expected immutable dail num %d, got %d", want, got)
		}
	}
}

func TestIgnoreDownEndpoints(t *testing.T) {
	var endpoints atomicMap
	endpoints.Store(prefixMap{"dns-not-found", "dial-timeout", "dial-refused", "up"})

	searcher := &HorizontalSearcher{
		Map: &endpoints,
		Dial: func(endpoint string) zoekt.Streamer {
			var client *FakeSearcher
			switch endpoint {
			case "dns-not-found":
				err := &net.DNSError{
					Err:        "no such host",
					Name:       "down",
					IsNotFound: true,
				}
				client = &FakeSearcher{
					SearchError: err,
					ListError:   err,
				}
			case "dial-timeout":
				// dial tcp 10.164.42.39:6070: i/o timeout
				err := &net.OpError{
					Op:   "dial",
					Net:  "tcp",
					Addr: fakeAddr("10.164.42.39:6070"),
					Err:  &timeoutError{},
				}
				client = &FakeSearcher{
					SearchError: err,
					ListError:   err,
				}
			case "dial-refused":
				// dial tcp 10.164.51.47:6070: connect: connection refused
				err := &net.OpError{
					Op:   "dial",
					Net:  "tcp",
					Addr: fakeAddr("10.164.51.47:6070"),
					Err:  errors.New("connect: connection refused"),
				}
				client = &FakeSearcher{
					SearchError: err,
					ListError:   err,
				}
			case "up":
				var rle zoekt.RepoListEntry
				rle.Repository.Name = "repo"

				client = &FakeSearcher{
					Result: &zoekt.SearchResult{
						Files: []zoekt.FileMatch{{
							Repository: "repo",
						}},
					},
					Repos: []*zoekt.RepoListEntry{&rle},
				}
			case "error":
				client = &FakeSearcher{
					SearchError: errors.New("boom"),
					ListError:   errors.New("boom"),
				}
			}

			return NewMeteredSearcher(endpoint, client)
		},
	}
	defer searcher.Close()

	sr, err := searcher.Search(context.Background(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sr.Files) == 0 {
		t.Fatal("Search: expected results")
	}

	rle, err := searcher.List(context.Background(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(rle.Repos) == 0 {
		t.Fatal("List: expected results")
	}

	// now test we do return errors if they occur
	endpoints.Store(prefixMap{"dns-not-found", "up", "error"})
	_, err = searcher.Search(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("Search: expected error")
	}

	_, err = searcher.List(context.Background(), nil, nil)
	if err == nil {
		t.Fatal("List: expected error")
	}
}

func TestResultQueueSettingsFromConfig(t *testing.T) {
	queueDepth := 96

	cases := []struct {
		name                   string
		siteConfig             schema.SiteConfiguration
		wantMaxQueueDepth      int
		wantMaxReorderDuration time.Duration
	}{
		{
			name:              "defaults",
			siteConfig:        schema.SiteConfiguration{},
			wantMaxQueueDepth: 24,
		},
		{
			name: "MaxReorderDurationMS",
			siteConfig: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{Ranking: &schema.Ranking{
				MaxReorderDurationMS: 5,
			}}},
			wantMaxQueueDepth:      24,
			wantMaxReorderDuration: 5 * time.Millisecond,
		},
		{

			name: "MaxReorderQueueSize",
			siteConfig: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{Ranking: &schema.Ranking{
				MaxReorderQueueSize: &queueDepth,
			}}},
			wantMaxQueueDepth: 96,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			haveMaxQueueDepth, haveMaxReorderDuration := resultQueueSettingsFromConfig(tt.siteConfig)

			if haveMaxQueueDepth != tt.wantMaxQueueDepth {
				t.Fatalf("want %d, got %d", tt.wantMaxQueueDepth, haveMaxQueueDepth)
			}

			if haveMaxReorderDuration != tt.wantMaxReorderDuration {
				t.Fatalf("want %d, got %d", tt.wantMaxReorderDuration, haveMaxReorderDuration)
			}
		})
	}
}

// implements net.Addr
type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }

type timeoutError struct{}

func (e *timeoutError) Error() string { return "i/o timeout" }
func (e *timeoutError) Timeout() bool { return true }

func TestDedupper(t *testing.T) {
	parse := func(s string) []zoekt.FileMatch {
		t.Helper()
		var fms []zoekt.FileMatch
		for _, t := range strings.Split(s, " ") {
			if t == "" {
				continue
			}
			parts := strings.Split(t, ":")
			fms = append(fms, zoekt.FileMatch{
				Repository: parts[0],
				FileName:   parts[1],
			})
		}
		return fms
	}
	cases := []struct {
		name    string
		matches []string
		want    string
	}{{
		name: "empty",
		matches: []string{
			"zoekt-0 ",
		},
		want: "",
	}, {
		name: "one",
		matches: []string{
			"zoekt-0 r1:a r1:a r1:b r2:a",
		},
		want: "r1:a r1:a r1:b r2:a",
	}, {
		name: "some dups",
		matches: []string{
			"zoekt-0 r1:a r1:a r1:b r2:a",
			"zoekt-1 r1:c r1:c r3:a",
		},
		want: "r1:a r1:a r1:b r2:a r3:a",
	}, {
		name: "no dups",
		matches: []string{
			"zoekt-0 r1:a r1:a r1:b r2:a",
			"zoekt-1 r4:c r4:c r5:a",
		},
		want: "r1:a r1:a r1:b r2:a r4:c r4:c r5:a",
	}, {
		name: "shuffled",
		matches: []string{
			"zoekt-0 r1:a r2:a r1:a r1:b",
			"zoekt-1 r1:c r3:a r1:c",
		},
		want: "r1:a r2:a r1:a r1:b r3:a",
	}, {
		name: "some dups multi event",
		matches: []string{
			"zoekt-0 r1:a r1:a",
			"zoekt-1 r1:c r1:c r3:a",
			"zoekt-0 r1:b r2:a",
		},
		want: "r1:a r1:a r3:a r1:b r2:a",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := dedupper{}
			var got []zoekt.FileMatch
			for _, s := range tc.matches {
				parts := strings.SplitN(s, " ", 2)
				endpoint := parts[0]
				fms := parse(parts[1])
				got = append(got, d.Dedup(endpoint, fms)...)
			}

			want := parse(tc.want)
			if !cmp.Equal(want, got, cmpopts.EquateEmpty()) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}

func BenchmarkDedup(b *testing.B) {
	nRepos := 100
	nMatchPerRepo := 50
	// primes to avoid the need of dedup most of the time :)
	shardStrides := []int{7, 5, 3, 2, 1}

	shardsOrig := [][]zoekt.FileMatch{}
	for _, stride := range shardStrides {
		shard := []zoekt.FileMatch{}
		for i := stride; i <= nRepos; i += stride {
			repo := fmt.Sprintf("repo-%d", i)
			for j := 0; j < nMatchPerRepo; j++ {
				path := fmt.Sprintf("%d.go", j)
				shard = append(shard, zoekt.FileMatch{
					Repository: repo,
					FileName:   path,
				})
			}
		}
		shardsOrig = append(shardsOrig, shard)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		// Create copy since we mutate the input in Deddup
		b.StopTimer()
		shards := make([][]zoekt.FileMatch, 0, len(shardsOrig))
		for _, shard := range shardsOrig {
			shards = append(shards, append([]zoekt.FileMatch{}, shard...))
		}
		b.StartTimer()

		d := dedupper{}
		for clientID, shard := range shards {
			_ = d.Dedup(strconv.Itoa(clientID), shard)
		}
	}
}

func backgroundSearch(searcher zoekt.Searcher) func(t *testing.T) {
	done := make(chan struct{})
	errC := make(chan error)
	go func() {
		for {
			_, err := searcher.Search(context.Background(), nil, nil)
			if err != nil {
				errC <- err
				return
			}
			_, err = searcher.List(context.Background(), nil, nil)
			if err != nil {
				errC <- err
				return
			}

			select {
			case <-done:
				errC <- err
				return
			default:
			}
		}
	}()

	return func(t *testing.T) {
		t.Helper()
		close(done)
		if err := <-errC; err != nil {
			t.Error("concurrent search failed: ", err)
		}
	}
}

type atomicMap struct {
	atomic.Value
}

func (m *atomicMap) Endpoints() ([]string, error) {
	return m.Value.Load().(EndpointMap).Endpoints()
}

func (m *atomicMap) Get(key string) (string, error) {
	return m.Value.Load().(EndpointMap).Get(key)
}
