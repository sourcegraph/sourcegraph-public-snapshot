package docsite

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"os"

	"github.com/sourcegraph/docsite"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

// Site is the documentation site.
var Site = docsite.Site{
	Content: nonVersionedFileSystem{content},
	Base:    &url.URL{Path: "/help/"},
}

type nonVersionedFileSystem struct{ http.FileSystem }

func (fs nonVersionedFileSystem) OpenVersion(_ context.Context, version string) (http.FileSystem, error) {
	if version != "" {
		return nil, errors.New("content versioning is not supported")
	}
	return fs.FileSystem, nil
}

// indexTemplate renders the HTML for the page index.
//
// It is rendered on the server (instead of the client) for 2 reasons: (1) because we want it to be
// identical (or nearly so) to that on docs.sourcegraph.com, and (2) it is difficult to represent
// tree data structures in GraphQL.
//
// NOTE: This is copied from the https://github.com/sourcegraph/docs.sourcegraph.com repository and
// should stay in sync.
var indexTemplate = template.Must(template.New("").Parse(`
{{define "index"}}
	{{with (or (and (eq (len .Doc.Tree) 1) (index .Doc.Tree 0).Children) .Doc.Tree)}}
		<h4 class="visible-sm">{{$.Doc.Title}}</h4>
		<h4 class="visible-lg">On this page:</h4>
		<ul>{{template "doc_nav" .}}</ul>
	{{end}}
	<a class="edit-btn" href="https://github.com/sourcegraph/sourcegraph/edit/master/doc/{{.FilePath}}">Edit this page</a>
{{end}}
{{define "doc_nav"}}
	{{range .}}
		<li>
			<a href="{{.URL}}">{{.Title}}</a>
			{{with .Children}}
				<ul>
					{{template "doc_nav" .}}
				</ul>
			{{end}}
	{{end}}
{{end}}
`))

// Register the resolver for the GraphQL field Query.docSitePage.
func init() {
	graphqlbackend.DocSitePageResolver = func(ctx context.Context, args graphqlbackend.DocSitePageArgs) (graphqlbackend.DocSitePage, error) {
		page, err := Site.ResolveContentPage(ctx, "", args.Path)
		if os.IsNotExist(err) {
			return nil, nil
		} else if err != nil {
			return nil, err
		}

		// See the comment on indexTemplate for why we render this on the server.
		var indexHTML bytes.Buffer
		if err := indexTemplate.ExecuteTemplate(&indexHTML, "index", page); err != nil {
			return nil, err
		}

		return &docSitePage{
			title:       page.Doc.Title,
			contentHTML: string(page.Doc.HTML),
			indexHTML:   indexHTML.String(),
			filePath:    page.FilePath,
		}, nil
	}
}

// docSitePage implements the GraphQL type DocSite.
type docSitePage struct {
	title       string
	contentHTML string
	indexHTML   string
	filePath    string
}

func (r *docSitePage) Title() string       { return r.title }
func (r *docSitePage) ContentHTML() string { return r.contentHTML }
func (r *docSitePage) IndexHTML() string   { return r.indexHTML }
func (r *docSitePage) FilePath() string    { return r.filePath }

var _ graphqlbackend.DocSitePage = &docSitePage{}
