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
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/regexp"
	"github.com/sourcegraph/zoekt"
	zjson "github.com/sourcegraph/zoekt/json"
	"github.com/sourcegraph/zoekt/query"
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
	"TrimTrailingNewline": func(s string) string {
		return strings.TrimSuffix(s, "\n")
	},
}

const defaultNumResults = 50

type Server struct {
	Searcher zoekt.Streamer

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

	// This should contain the following templates: "repolist"
	// (for the repo search result page), "result" for
	// the search results, "search" (for the opening page),
	// "box" for the search query input element and
	// "print" for the show file functionality.
	Top *template.Template

	repolist *template.Template
	search   *template.Template
	result   *template.Template
	print    *template.Template
	about    *template.Template
	robots   *template.Template

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
		"results":  &s.result,
		"print":    &s.print,
		"search":   &s.search,
		"repolist": &s.repolist,
		"about":    &s.about,
		"robots":   &s.robots,
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
		mux.HandleFunc("/robots.txt", s.serveRobots)
		mux.HandleFunc("/search", s.serveSearch)
		mux.HandleFunc("/", s.serveSearchBox)
		mux.HandleFunc("/about", s.serveAbout)
		mux.HandleFunc("/print", s.servePrint)
	}
	if s.RPC {
		mux.Handle("/api/", http.StripPrefix("/api", zjson.JSONServer(traceAwareSearcher{s.Searcher})))
	}

	mux.HandleFunc("/healthz", s.serveHealthz)

	return mux, nil
}

func (s *Server) serveHealthz(w http.ResponseWriter, r *http.Request) {
	q := &query.Const{Value: true}
	opts := &zoekt.SearchOptions{ShardMaxMatchCount: 1, TotalMaxMatchCount: 1, MaxDocDisplayCount: 1}

	result, err := s.Searcher.Search(r.Context(), q, opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("not ready: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(result)
}

func (s *Server) serveSearch(w http.ResponseWriter, r *http.Request) {
	result, err := s.serveSearchErr(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}

	qvals := r.URL.Query()
	if qvals.Get("format") == "json" {
		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(result)
		return
	}

	var buf bytes.Buffer
	if result.Repos != nil {
		err = s.repolist.Execute(&buf, &result.Repos)
	} else if result.Result != nil {
		err = s.result.Execute(&buf, &result.Result)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}
	w.Write(buf.Bytes())
}

func (s *Server) serveSearchErr(r *http.Request) (*ApiSearchResult, error) {
	qvals := r.URL.Query()

	debugScore, _ := strconv.ParseBool(qvals.Get("debug"))

	queryStr := qvals.Get("q")
	if queryStr == "" {
		return nil, fmt.Errorf("no query found")
	}

	q, err := query.Parse(queryStr)
	if err != nil {
		return nil, err
	}

	repoOnly := true
	query.VisitAtoms(q, func(q query.Q) {
		_, ok := q.(*query.Repo)
		repoOnly = repoOnly && ok
	})
	if repoOnly {
		repos, err := s.serveListReposErr(q, queryStr, r)
		if err == nil {
			return &ApiSearchResult{Repos: repos}, nil
		}
		return nil, err
	}

	if qt, ok := q.(*query.Type); ok && qt.Type == query.TypeRepo {
		repos, err := s.serveListReposErr(q, queryStr, r)
		if err == nil {
			return &ApiSearchResult{Repos: repos}, nil
		}
		return nil, err
	}

	numStr := qvals.Get("num")

	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		num = defaultNumResults
	}

	sOpts := zoekt.SearchOptions{
		MaxWallTime: 10 * time.Second,
	}

	numCtxLines := 0
	if ctxLinesStr := qvals.Get("ctx"); ctxLinesStr != "" {
		numCtxLines, err = strconv.Atoi(ctxLinesStr)
		if err != nil || numCtxLines < 0 || numCtxLines > 10 {
			return nil, fmt.Errorf("Number of context lines must be between 0 and 10")
		}
	}
	sOpts.NumContextLines = numCtxLines

	sOpts.SetDefaults()
	sOpts.MaxDocDisplayCount = num
	sOpts.DebugScore = debugScore

	ctx := r.Context()
	if err := zjson.CalculateDefaultSearchLimits(ctx, q, s.Searcher, &sOpts); err != nil {
		return nil, err
	}

	result, err := s.Searcher.Search(ctx, q, &sOpts)
	if err != nil {
		return nil, err
	}

	fileMatches, err := s.formatResults(result, queryStr, s.Print)
	if err != nil {
		return nil, err
	}

	res := ResultInput{
		Last: LastInput{
			Query:     queryStr,
			Num:       num,
			Ctx:       numCtxLines,
			AutoFocus: true,
		},
		Stats:       result.Stats,
		Query:       q.String(),
		QueryStr:    queryStr,
		FileMatches: fileMatches,
	}
	if res.Stats.Wait < res.Stats.Duration/10 {
		// Suppress queueing stats if they are neglible.
		res.Stats.Wait = 0
	}

	res.Last.Debug = debugScore
	return &ApiSearchResult{Result: &res}, nil
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
	if time.Since(s.lastStatsTS) > statsStaleNess {
		stats = nil
	}
	s.lastStatsMu.Unlock()

	if stats != nil {
		return stats, nil
	}

	repos, err := s.Searcher.List(ctx, &query.Const{Value: true}, nil)
	if err != nil {
		return nil, err
	}

	stats = &repos.Stats

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
		Uptime:  time.Since(s.startTime),
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
	_, _ = w.Write(buf.Bytes())
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
		Uptime:  time.Since(s.startTime),
	}

	var buf bytes.Buffer
	if err := s.about.Execute(&buf, &d); err != nil {
		return err
	}
	_, _ = w.Write(buf.Bytes())
	return nil
}

func (s *Server) serveAbout(w http.ResponseWriter, r *http.Request) {
	if err := s.serveAboutErr(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveRobotsErr(w http.ResponseWriter, r *http.Request) error {
	data := struct{}{}
	var buf bytes.Buffer
	if err := s.robots.Execute(&buf, &data); err != nil {
		return err
	}
	_, _ = w.Write(buf.Bytes())
	return nil
}

func (s *Server) serveRobots(w http.ResponseWriter, r *http.Request) {
	if err := s.serveRobotsErr(w, r); err != nil {
		http.Error(w, err.Error(), http.StatusTeapot)
	}
}

func (s *Server) serveListReposErr(q query.Q, qStr string, r *http.Request) (*RepoListInput, error) {
	ctx := r.Context()
	repos, err := s.Searcher.List(ctx, q, nil)
	if err != nil {
		return nil, err
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
	case "ram", "revram":
		sort.Slice(repos.Repos, func(i, j int) bool {
			return repos.Repos[i].Stats.IndexBytes < repos.Repos[j].Stats.IndexBytes
		})
	case "time", "revtime":
		sort.Slice(repos.Repos, func(i, j int) bool {
			return repos.Repos[i].IndexMetadata.IndexTime.Before(
				repos.Repos[j].IndexMetadata.IndexTime)
		})
	default:
		return nil, fmt.Errorf("got unknown sort key %q, allowed [rev]name, [rev]time, [rev]size", order)
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

	numStr := qvals.Get("num")
	num, err := strconv.Atoi(numStr)
	if err != nil || num <= 0 {
		num = 0
	}
	if num > 0 {
		if num > len(repos.Repos) {
			num = len(repos.Repos)
		}

		repos.Repos = repos.Repos[:num]
	}

	res := RepoListInput{
		Last: LastInput{
			Query:     qStr,
			Num:       num,
			AutoFocus: true,
		},
		Stats: aggregate,
	}

	for _, r := range repos.Repos {
		t := s.getTemplate(r.Repository.CommitURLTemplate)

		repo := Repository{
			Name:       r.Repository.Name,
			URL:        r.Repository.URL,
			IndexTime:  r.IndexMetadata.IndexTime,
			Size:       r.Stats.ContentBytes,
			MemorySize: r.Stats.IndexBytes,
			Files:      int64(r.Stats.Documents),
		}
		for _, b := range r.Repository.Branches {
			var buf bytes.Buffer
			if err := t.Execute(&buf, b); err != nil {
				return nil, err
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
	return &res, nil
}

func (s *Server) servePrintErr(w http.ResponseWriter, r *http.Request) error {
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

	repoRe, err := regexp.Compile("^" + regexp.QuoteMeta(repoStr) + "$")
	if err != nil {
		return err
	}

	qs := []query.Q{
		&query.Regexp{Regexp: re, FileName: true, CaseSensitive: true},
		&query.Repo{Regexp: repoRe},
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

	_, _ = w.Write(buf.Bytes())
	return nil
}
