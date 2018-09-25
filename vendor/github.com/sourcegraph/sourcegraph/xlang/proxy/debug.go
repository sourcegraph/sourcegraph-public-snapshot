package proxy

import (
	"container/ring"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// DebugHandler is an HTTP handler which allows you to explore the active
// connections in a Proxy
type DebugHandler struct {
	Proxy *Proxy
}

func (h *DebugHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// build list of servers so we don't need to gain server locks while
	// holding the proxy lock.
	h.Proxy.mu.Lock()
	servers := make([]*serverProxyConn, 0, len(h.Proxy.servers))
	for c := range h.Proxy.servers {
		servers = append(servers, c)
	}
	h.Proxy.mu.Unlock()

	// build list of recently closed servers so we don't need to hold the
	// lock while computing summaries.
	closedServers := make([]*recentlyClosedItem, 0, recentlyClosedMaxItems)
	recentlyClosedMu.Lock()
	recentlyClosed.Do(func(x interface{}) {
		if x == nil {
			return
		}
		closedServers = append(closedServers, x.(*recentlyClosedItem))
	})
	recentlyClosedMu.Unlock()

	var data []*modeSummary
	{
		modes := make(map[string][]*serverSummary)
		closed := make(map[string][]*serverSummary)
		for _, c := range servers {
			modes[c.id.mode] = append(modes[c.id.mode], newServerSummary(c.id, c.Stats()))
		}
		for _, c := range closedServers {
			closed[c.id.mode] = append(closed[c.id.mode], newServerSummary(c.id, c.stats))
		}
		for mode := range modes {
			sort.Sort(byAge(modes[mode]))
			sort.Sort(byIdle(closed[mode]))
			data = append(data, &modeSummary{
				Name:      mode,
				TotalOpen: len(modes[mode]),
				Servers:   modes[mode],
				Closed:    closed[mode],
			})
		}
		sort.Sort(byTotalOpen(data))
	}

	err := debugTmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type serverSummary struct {
	ID         string
	TotalCount int
	Age        time.Duration
	Idle       time.Duration
	Counts     string
	Errors     string
}

type modeSummary struct {
	Name      string
	TotalOpen int
	Servers   []*serverSummary
	Closed    []*serverSummary
}

type countSummary struct {
	Method string
	Count  int
}

var debugTmpl = template.Must(template.New("").Parse(`
<html>
<head>
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" integrity="sha384-1q8mTJOASx8j1Au+a5WDVnPi2lkFfwwEAa8hDDdjZlpLegxhjVME1fgjWPGmkzs7" crossorigin="anonymous">
  <link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap-theme.min.css" integrity="sha384-fLW2N01lMqjakBkx3l/M9EahuwpSfeNvV63J5ezn3uZzapT0u7EYsXMjQV+0En5r" crossorigin="anonymous">
</head>
<body>
<div class="container">

<h1>LSP-Proxy current connections</h1>

<h2>Mode Summaries</h2>
<table class="table table-condensed table-hover">
<tr><th>Mode</th><th>TotalOpen</th></tr>
{{range .}}<tr><td>{{.Name}}</td><td>{{.TotalOpen}}</td></tr>{{end}}
</table>
</p>

<h2>Modes</h2>
{{range .}}
  <h3>{{.Name}} ({{.TotalOpen}} open)</h3>
  <table class="table table-condensed table-hover">
  <tr><th>ServerID</th><th>TotalCount</th><th>Age</th><th>Idle</th><th>Counts</th><th>Errors</th></tr>
  {{range .Servers}}<tr><td>{{.ID}}</td><td>{{.TotalCount}}</td><td>{{.Age}}</td><td>{{.Idle}}</td><td>{{.Counts}}</td><td>{{.Errors}}</td></tr>{{end}}
  </table>
  {{if .Closed}}
  <h4>{{.Name}} Recently Closed</h4>
  <table class="table table-condensed table-hover">
  <tr><th>ServerID</th><th>TotalCount</th><th>Age</th><th>Idle</th><th>Counts</th><th>Errors</th></tr>
  {{range .Closed}}<tr><td>{{.ID}}</td><td>{{.TotalCount}}</td><td>{{.Age}}</td><td>{{.Idle}}</td><td>{{.Counts}}</td><td>{{.Errors}}</td></tr>{{end}}
  </table>
  {{end}}
{{end}}

</div>
</body>
</html>
`))

const recentlyClosedMaxItems = 100

var (
	recentlyClosedMu sync.Mutex
	recentlyClosed   = ring.New(recentlyClosedMaxItems)
)

type recentlyClosedItem struct {
	id    serverID
	stats serverProxyConnStats
}

// recordClosedServerConn adds the server stats to the recently closed list
// shown on the debug page.
func recordClosedServerConn(id serverID, stats serverProxyConnStats) {
	item := &recentlyClosedItem{
		id:    id,
		stats: stats,
	}
	recentlyClosedMu.Lock()
	recentlyClosed.Value = item
	recentlyClosed = recentlyClosed.Next()
	recentlyClosedMu.Unlock()
}

func newServerSummary(id serverID, stats serverProxyConnStats) *serverSummary {
	return &serverSummary{
		ID:         id.rootURI.String(),
		TotalCount: stats.TotalCount,
		Age:        time.Since(stats.Created),
		Idle:       time.Since(stats.Last),
		Counts:     summariseCounts(stats.Counts),
		Errors:     summariseCounts(stats.ErrorCounts),
	}
}

type byTotalOpen []*modeSummary

func (p byTotalOpen) Len() int      { return len(p) }
func (p byTotalOpen) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p byTotalOpen) Less(i, j int) bool {
	if p[i].TotalOpen == p[j].TotalOpen {
		return p[i].Name < p[j].Name
	}
	// > since we want descending sort
	return p[i].TotalOpen > p[j].TotalOpen
}

type byAge []*serverSummary

func (p byAge) Len() int           { return len(p) }
func (p byAge) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p byAge) Less(i, j int) bool { return p[i].Age > p[j].Age }

type byIdle []*serverSummary

func (p byIdle) Len() int           { return len(p) }
func (p byIdle) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p byIdle) Less(i, j int) bool { return p[i].Idle > p[j].Idle }

type byCount []countSummary

func (p byCount) Len() int           { return len(p) }
func (p byCount) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p byCount) Less(i, j int) bool { return p[i].Count > p[j].Count }

func summariseCounts(counts map[string]int) string {
	cs := make([]countSummary, 0, len(counts))
	for method, count := range counts {
		cs = append(cs, countSummary{
			Method: shortenMethodName(method),
			Count:  count,
		})
	}
	sort.Sort(byCount(cs))
	p := make([]string, len(cs))
	for i, c := range cs {
		p[i] = fmt.Sprintf("%s=%d", c.Method, c.Count)
	}
	return strings.Join(p, " ")
}

// shortenMethodNameRe finds the first character of each word when using
// camelcase. It special cases x since that is a common prefix for
// experimental extensions.
var shortenMethodNameRe = regexp.MustCompile(`(^x?[a-wyz]|[A-Z]|/x?[a-wyz])`)

// shortenMethodName shorterns an LSP method name for display purposes.
// eg textDocument/hover becomes td/h
func shortenMethodName(method string) string {
	parts := shortenMethodNameRe.FindAllString(method, -1)
	return strings.ToLower(strings.Join(parts, ""))
}
