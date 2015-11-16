package tracker

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"

	"github.com/shurcooL/htmlg"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"src.sourcegraph.com/apps/tracker/issues"
)

// TODO: Factor out somehow...
var tabsTmpl = template.Must(template.New("").Parse(`
{{define "open-issue-count"}}<span><span style="margin-right: 4px;" class="octicon octicon-issue-opened"></span>{{.OpenCount}} Open</span>{{end}}
{{define "closed-issue-count"}}<span><span style="margin-right: 4px;" class="octicon octicon-check"></span>{{.ClosedCount}} Closed</span>{{end}}
`))

const (
	queryKeyState = "state"
)

// TODO: Reorganize and deduplicate.
func tab(query url.Values) (issues.State, error) {
	switch query.Get(queryKeyState) {
	case "":
		return issues.OpenState, nil
	case string(issues.ClosedState):
		return issues.ClosedState, nil
	default:
		return "", fmt.Errorf("unsupported queryKeyState value: %q", query.Get(queryKeyState))
	}
}

// tabs renders the html for <nav> element with tab header links.
func tabs(s *state, path string, rawQuery string) (template.HTML, error) {
	query, _ := url.ParseQuery(rawQuery)

	selectedTab := query.Get(queryKeyState)

	var ns []*html.Node

	for _, tab := range []struct {
		id           string
		templateName string
	}{
		{id: "", templateName: "open-issue-count"},
		{id: string(issues.ClosedState), templateName: "closed-issue-count"},
	} {
		li := &html.Node{Type: html.ElementNode, Data: atom.Li.String()}
		a := &html.Node{Type: html.ElementNode, Data: atom.A.String()}
		if tab.id == selectedTab {
			li.Attr = []html.Attribute{{Key: atom.Class.String(), Val: "active"}}
		} else {
			q := query
			if tab.id == "" {
				q.Del(queryKeyState)
			} else {
				q.Set(queryKeyState, tab.id)
			}
			u := url.URL{
				Path:     path,
				RawQuery: q.Encode(),
			}
			a.Attr = []html.Attribute{
				{Key: atom.Href.String(), Val: u.String()},
			}
		}
		// TODO: This is horribly inefficient... :o
		var buf bytes.Buffer
		err := tabsTmpl.ExecuteTemplate(&buf, tab.templateName, s)
		if err != nil {
			return "", err
		}
		tmplNode, err := html.Parse(&buf)
		if err != nil {
			return "", err
		}
		a.AppendChild(tmplNode)
		li.AppendChild(a)
		ns = append(ns, li)
	}

	return htmlg.Render(ns...), nil
}
