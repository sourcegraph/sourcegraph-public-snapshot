// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"regexp"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/google/zoekt"
	"github.com/google/zoekt/query"
	"github.com/google/zoekt/rpc"
)

var Funcmap = template.FuncMap{
	"Inc": func(orig int) int {
		return orig + 1
	},
	"More": func(orig int) int {
		return orig * 3
	},
	"HumanUnit": func(orig int64) string {
		b := orig
		suffix := ""
		if orig > 10*(1<<30) {
			suffix = "G"
			b = orig / (1 << 30)
		} else if orig > 10*(1<<20) {
			suffix = "M"
			b = orig / (1 << 20)
		} else if orig > 10*(1<<10) {
			suffix = "K"
			b = orig / (1 << 10)
		}

		return fmt.Sprintf("%d%s", b, suffix)
	},
	"LimitPre": func(limit int, pre string) string {
		if len(pre) < limit {
			return pre
		}
		return fmt.Sprintf("...(%d bytes skipped)...%s", len(pre)-limit, pre[len(pre)-limit:])
	},
	"LimitPost": func(limit int, post string) string {
		if len(post) < limit {
			return post
		}
		return fmt.Sprintf("%s...(%d bytes skipped)...", post[:limit], len(post)-limit)
	},
}

const defaultNumResults = 50

type Server struct {
	Searcher zoekt.Searcher

	// Serve HTML interface
	HTML bool

	// Serve RPC
	RPC bool

	// If set, show files from the index.
	Print bool

	// Version string for this server.
	Version string

	// Depending on the Host header, add a query to the entry
	// page. For example, when serving on "search.myproject.org"
	// we could add "r:myproject" automatically.  This allows a
	// single instance to serve as search engine for multiple
	// domains.
	HostCustomQueries map[string]string

	// This should contain the following templates: "didyoumean"
	// (for suggestions), "repolist" (for the repo search result
	// page), "result" for the search results, "search" (for the
	// opening page), "box" for the search query input element and
	// "print" for the show file functionality.
	Top *template.Template

	didYouMean *template.Template
	repolist   *template.Template
	search     *template.Template
	result     *template.Template
	print      *template.Template
	about      *template.Template

	startTime time.Time

	templateMu    sync.Mutex
	templateCache map[string]*template.Template

	lastStatsMu sync.Mutex
	lastStats   *zoekt.RepoStats
	lastStatsTS time.Time
}

func (s *Server) getTemplate(str string) *template.Template {
	s.templateMu.Lock()
	defer s.templateMu.Unlock()
	t := s.templateCache[str]
	if t != nil {
		return t
	}

	t, err := template.New("cache").Parse(str)
	if err != nil {
		log.Printf("template parse error: %v", err)
		t = template.Must(template.New("empty").Parse(""))
	}
	s.templateCache[str] = t
	return t
}

func NewMux(s *Server) (*http.ServeMux, error) {
	s.print = s.Top.Lookup("print")
	if s.print == nil {
		return nil, fmt.Errorf("missing template 'print'")
	}

	for k, v := range map[string]**template.Template{
		"didyoumean": &s.didYouMean,
		"results":    &s.result,
		"print":      &s.print,
		"search":     &s.search,
		"repolist":   &s.repolist,
		"about":      &s.about,
	} {
		*v = s.Top.Lookup(k)
		if *v == nil {
			return nil, fmt.Errorf("missing template %q", k)
		}
	}

	s.templateCache = map[string]*template.Template{}
	s.startTime = time.Now()

	mux := http.NewServeMux()

	if s.HTML {
		mux.HandleFunc("/search", s.serveSearch)
		mux.HandleFunc("/", s.serveSearchBox)
		mux.HandleFunc("/about", s.serveAbout)
	}
	if s.RPC {
		mux.Handle(rpc.DefaultRPCPath, rpc.Server(s.Searcher)) // /rpc
	}
	if s.Print {
		mux.HandleFunc("/print", s.servePrint)
	}
	return mux, nil
}

func (s *Server) serveSearch(w http.ResponseWriter, r *http.Request) {
	err := s.serveSearchErr(w, r)

	if suggest, ok := err.(*query.SuggestQueryError); ok {
		var buf bytes.Buffer
		if err := s.didYouMean.Execute(&buf, suggest); err != nil {
			http.Error(w, err.Error(), http.StatusTeapot)
		}

		w.Write(buf.Bytes())
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveSearchErr(w http.ResponseWriter, r *http.Request) error {
	qvals := r.URL.Query()
	queryStr := qvals.Get("q")
	if queryStr == "" {
		return fmt.Errorf("no query found")
	}

	log.Printf("got query %q", queryStr)
	q, err := query.Parse(queryStr)
	if err != nil {
		return err
	}

	repoOnly := true
	query.VisitAtoms(q, func(q query.Q) {
		_, ok := q.(*query.Repo)
		repoOnly = repoOnly && ok
	})
	if repoOnly {
		return s.serveListReposErr(q, queryStr, w, r)
	}

	numStr := qvals.Get("num")

	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		num = defaultNumResults
	}

	sOpts := zoekt.SearchOptions{
		MaxWallTime: 10 * time.Second,
	}

	sOpts.SetDefaults()

	ctx := r.Context()
	if result, err := s.Searcher.Search(ctx, q, &zoekt.SearchOptions{EstimateDocCount: true}); err != nil {
		return err
	} else if numdocs := result.ShardFilesConsidered; numdocs > 10000 {
		// If the search touches many shards and many files, we
		// have to limit the number of matches.  This setting
		// is based on the number of documents eligible after
		// considering reponames, so large repos (both
		// android, chromium are about 500k files) aren't
		// covered fairly.

		// 10k docs, 50 num -> max match = (250 + 250 / 10)
		sOpts.ShardMaxMatchCount = num*5 + (5*num)/(numdocs/1000)

		// 10k docs, 50 num -> max important match = 4
		sOpts.ShardMaxImportantMatch = num/20 + num/(numdocs/500)
	} else {
		// Virtually no limits for a small corpus; important
		// matches are just as expensive as normal matches.
		n := numdocs + num*100
		sOpts.ShardMaxImportantMatch = n
		sOpts.ShardMaxMatchCount = n
		sOpts.TotalMaxMatchCount = n
		sOpts.TotalMaxImportantMatch = n
	}
	sOpts.MaxDocDisplayCount = num

	result, err := s.Searcher.Search(ctx, q, &sOpts)
	if err != nil {
		return err
	}

	fileMatches, err := s.formatResults(result, queryStr, s.Print)
	if err != nil {
		return err
	}

	res := ResultInput{
		Last: LastInput{
			Query:     queryStr,
			Num:       num,
			AutoFocus: true,
		},
		Stats:         result.Stats,
		Query:         q.String(),
		QueryStr:      queryStr,
		SearchOptions: sOpts.String(),
		FileMatches:   fileMatches,
	}
	if res.Stats.Wait < res.Stats.Duration/10 {
		// Suppress queueing stats if they are neglible.
		res.Stats.Wait = 0
	}

	var buf bytes.Buffer
	if err := s.result.Execute(&buf, &res); err != nil {
		return err
	}

	w.Write(buf.Bytes())
	return nil
}

func (s *Server) servePrint(w http.ResponseWriter, r *http.Request) {
	err := s.servePrintErr(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

const statsStaleNess = 30 * time.Second

func (s *Server) fetchStats(ctx context.Context) (*zoekt.RepoStats, error) {
	s.lastStatsMu.Lock()
	stats := s.lastStats
	if time.Now().Sub(s.lastStatsTS) > statsStaleNess {
		stats = nil
	}
	s.lastStatsMu.Unlock()

	if stats != nil {
		return stats, nil
	}

	repos, err := s.Searcher.List(ctx, &query.Const{Value: true})
	if err != nil {
		return nil, err
	}

	stats = &zoekt.RepoStats{}
	names := map[string]struct{}{}
	for _, r := range repos.Repos {
		stats.Add(&r.Stats)
		names[r.Repository.Name] = struct{}{}
	}
	stats.Repos = len(names)

	s.lastStatsMu.Lock()
	s.lastStatsTS = time.Now()
	s.lastStats = stats
	s.lastStatsMu.Unlock()

	return stats, nil
}

func (s *Server) serveSearchBoxErr(w http.ResponseWriter, r *http.Request) error {
	stats, err := s.fetchStats(r.Context())
	if err != nil {
		return err
	}
	d := SearchBoxInput{
		Last: LastInput{
			Num:       defaultNumResults,
			AutoFocus: true,
		},
		Stats:   stats,
		Version: s.Version,
		Uptime:  time.Now().Sub(s.startTime),
	}

	d.Last.Query = r.URL.Query().Get("q")
	if d.Last.Query == "" {
		custom := s.HostCustomQueries[r.Host]
		if custom == "" {
			host, _, _ := net.SplitHostPort(r.Host)
			custom = s.HostCustomQueries[host]
		}

		if custom != "" {
			d.Last.Query = custom + " "
		}
	}

	var buf bytes.Buffer
	if err := s.search.Execute(&buf, &d); err != nil {
		return err
	}
	w.Write(buf.Bytes())
	return nil
}

func (s *Server) serveSearchBox(w http.ResponseWriter, r *http.Request) {
	if err := s.serveSearchBoxErr(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveAboutErr(w http.ResponseWriter, r *http.Request) error {
	stats, err := s.fetchStats(r.Context())
	if err != nil {
		return err
	}

	d := SearchBoxInput{
		Stats:   stats,
		Version: s.Version,
		Uptime:  time.Now().Sub(s.startTime),
	}

	var buf bytes.Buffer
	if err := s.about.Execute(&buf, &d); err != nil {
		return err
	}
	w.Write(buf.Bytes())
	return nil
}

func (s *Server) serveAbout(w http.ResponseWriter, r *http.Request) {
	if err := s.serveAboutErr(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveListReposErr(q query.Q, qStr string, w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	repos, err := s.Searcher.List(ctx, q)
	if err != nil {
		return err
	}

	qvals := r.URL.Query()
	order := qvals.Get("order")
	switch order {
	case "", "name", "revname":
		sort.Slice(repos.Repos, func(i, j int) bool {
			return strings.Compare(repos.Repos[i].Repository.Name, repos.Repos[j].Repository.Name) < 0
		})
	case "size", "revsize":
		sort.Slice(repos.Repos, func(i, j int) bool {
			return repos.Repos[i].Stats.ContentBytes < repos.Repos[j].Stats.ContentBytes
		})
	case "time", "revtime":
		sort.Slice(repos.Repos, func(i, j int) bool {
			return repos.Repos[i].IndexMetadata.IndexTime.Before(
				repos.Repos[j].IndexMetadata.IndexTime)
		})
	default:
		return fmt.Errorf("got unknown sort key %q, allowed [rev]name, [rev]time, [rev]size", order)
	}
	if strings.HasPrefix(order, "rev") {
		for i, j := 0, len(repos.Repos)-1; i < j; {
			repos.Repos[i], repos.Repos[j] = repos.Repos[j], repos.Repos[i]
			i++
			j--

		}
	}

	aggregate := zoekt.RepoStats{
		Repos: len(repos.Repos),
	}
	for _, s := range repos.Repos {
		aggregate.Add(&s.Stats)
	}
	res := RepoListInput{
		Last: LastInput{
			Query:     qStr,
			AutoFocus: true,
		},
		Stats: aggregate,
	}

	for _, r := range repos.Repos {
		t := s.getTemplate(r.Repository.CommitURLTemplate)

		repo := Repository{
			Name:      r.Repository.Name,
			URL:       r.Repository.URL,
			IndexTime: r.IndexMetadata.IndexTime,
			Size:      r.Stats.ContentBytes,
			Files:     int64(r.Stats.Documents),
		}
		for _, b := range r.Repository.Branches {
			var buf bytes.Buffer
			if err := t.Execute(&buf, b); err != nil {
				return err
			}
			repo.Branches = append(repo.Branches,
				Branch{
					Name:    b.Name,
					Version: b.Version,
					URL:     buf.String(),
				})
		}
		res.Repos = append(res.Repos, repo)
	}

	var buf bytes.Buffer
	if err := s.repolist.Execute(&buf, &res); err != nil {
		return err
	}

	w.Write(buf.Bytes())
	return nil
}

func (s *Server) servePrintErr(w http.ResponseWriter, r *http.Request) error {
	if !s.Print {
		return fmt.Errorf("no printing template defined.")
	}

	qvals := r.URL.Query()
	fileStr := qvals.Get("f")
	repoStr := qvals.Get("r")
	queryStr := qvals.Get("q")
	numStr := qvals.Get("num")
	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		num = defaultNumResults
	}

	re, err := syntax.Parse("^"+regexp.QuoteMeta(fileStr)+"$", 0)
	if err != nil {
		return err
	}
	qs := []query.Q{
		&query.Regexp{Regexp: re, FileName: true, CaseSensitive: true},
		&query.Repo{Pattern: repoStr},
	}

	if branchStr := qvals.Get("b"); branchStr != "" {
		qs = append(qs, &query.Branch{Pattern: branchStr})
	}

	q := &query.And{Children: qs}

	sOpts := zoekt.SearchOptions{
		Whole: true,
	}

	ctx := r.Context()
	result, err := s.Searcher.Search(ctx, q, &sOpts)
	if err != nil {
		return err
	}

	if len(result.Files) != 1 {
		var ss []string
		for _, n := range result.Files {
			ss = append(ss, n.FileName)
		}
		return fmt.Errorf("ambiguous result: %v", ss)
	}

	f := result.Files[0]

	byteLines := bytes.Split(f.Content, []byte{'\n'})
	strLines := make([]string, 0, len(byteLines))
	for _, l := range byteLines {
		strLines = append(strLines, string(l))
	}

	d := PrintInput{
		Name:  f.FileName,
		Repo:  f.Repository,
		Lines: strLines,
		Last: LastInput{
			Query:     queryStr,
			Num:       num,
			AutoFocus: false,
		},
	}

	var buf bytes.Buffer
	if err := s.print.Execute(&buf, &d); err != nil {
		return err
	}

	w.Write(buf.Bytes())
	return nil
}
