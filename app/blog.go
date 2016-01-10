package app

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/util/htmlutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

// Number of blog posts to display per page on the index page and in the atom
// feed, Tumblr API has a limit of 20.
const blogPostsPerPage = 20

type blogPostMeta struct {
	Title  string
	Date   time.Time
	URL    string `json:"slug"`
	Author struct{ Login, Name string }
}

// swapOldMeta swaps this metadata with old JSON metadata (see blog_old_meta.go)
// for Tumblr posts that used to be Markdown posts and had their date
// or other metadata information lost.
func (m *blogPostMeta) swapOldMeta(p *Post) {
	postID := strconv.Itoa(p.ID)

	// Search for the alias.
	for oldSlug, newSlug := range blogSlugAliases {
		if postID == newSlug {
			if oldMeta, ok := blogOldMetaData[oldSlug]; ok {
				*m = *oldMeta
			}
		}
	}

	// Always adopt the newest Tumblr post URL instead of using older legacy ones.
	u, err := url.Parse(p.PostURL)
	if err != nil {
		// Fallback to just the post ID itself.
		m.URL = postID
		return
	}
	m.URL = strings.TrimPrefix(u.Path, "/post/")
}

// newBlogPostMeta returns a blog post's metadata structure given the post
// itself. It also swaps the metadata out with old meta-data if needed.
func newBlogPostMeta(p *Post) *blogPostMeta {
	m := &blogPostMeta{
		Title: p.Title,
		Date:  time.Unix(int64(p.Timestamp), 0),

		// Tumblr API doesn't give us anything except the author's Tumblr username,
		// we assume the best case which is that the author's username on Tumblr
		// matches their GitHub username (i.e. sourcegraph.com/username) -- and that
		// displaying their username in place of their real name is OK.
		Author: struct{ Login, Name string }{p.PostAuthor, p.PostAuthor},
	}
	m.swapOldMeta(p)
	return m
}

var tumblrBlog = &tumblr{
	Path:       "/blog",
	Blog:       "sourcegraph-com.tumblr.com",
	BlogTitle:  "Sourcegraph Blog",
	BlogBanner: `<h1 class="blog-title">Sourcegraph Blog</h1>`,
}

// blogFetch fetches a single page of the blog index from Tumblr, it returns the
// response and a list of built meta-data. When needed (see blog_old_meta.go) it
// properly uses the old meta-data (e.g. for the author name).
func blogFetch(page, perPage int) ([]*blogPostMeta, *PostsResponse, error) {
	// Query the posts.
	postsResp, err := tumblrBlog.Posts(PostsOpts{
		ID:              "",
		DisableSanitize: true,
		ListOptions: ListOptions{
			PerPage: perPage,
			Page:    page,
		},
		Type: typeText,
	})
	if err != nil {
		return nil, nil, err
	}

	// Create the meta-data structures.
	postsMeta := make([]*blogPostMeta, len(postsResp.Posts))
	for i, p := range postsResp.Posts {
		postsMeta[i] = newBlogPostMeta(p)
	}
	return postsMeta, postsResp, nil
}

// blogQueryOpts grabs the post options from the URL query schema and returns
// them. opt.PerPage is always equal to the blogPostsPerPage constant.
func blogQueryOpts(r *http.Request) (*PostsOpts, error) {
	var opt PostsOpts
	if err := schema.NewDecoder().Decode(&opt, r.URL.Query()); err != nil {
		return nil, err
	}
	opt.PerPage = blogPostsPerPage
	if opt.Page == 0 {
		opt.Page = 1
	}
	return &opt, nil
}

func serveBlogIndex(w http.ResponseWriter, r *http.Request) error {
	// Grab the post options from the URL query.
	opt, err := blogQueryOpts(r)
	if err != nil {
		return err
	}

	// Fetch the blog posts.
	posts, resp, err := blogFetch(opt.Page, opt.PerPage)
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("no such page"),
		}
	}

	return tmpl.Exec(r, w, "blog/index.html", http.StatusOK, nil, &struct {
		Posts       []*blogPostMeta
		RobotsIndex bool
		Response    *PostsResponse
		Limit       int
		Offset      int
		tmpl.Common
	}{
		Posts:       posts,
		RobotsIndex: true,
		Response:    resp,
		Limit:       opt.PerPage,
		Offset:      opt.PerPage * (opt.Page - 1),
	})
}

func serveBlogIndexAtom(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	// Grab the post options from the URL query.
	opt, err := blogQueryOpts(r)
	if err != nil {
		return err
	}

	// Fetch the blog posts.
	posts, resp, err := blogFetch(opt.Page, opt.PerPage)
	if err != nil {
		return err
	}
	if len(posts) == 0 {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("no such page"),
		}
	}

	type TimeStr string

	var Time = func(t time.Time) TimeStr {
		return TimeStr(t.Format("2006-01-02T15:04:05-07:00"))
	}

	type Link struct {
		Rel  string `xml:"rel,attr"`
		Href string `xml:"href,attr"`
		Type string `xml:"type,attr"`
	}

	type Person struct {
		Name     string `xml:"name"`
		URI      string `xml:"uri,omitempty"`
		Email    string `xml:"email,omitempty"`
		InnerXML string `xml:",innerxml"`
	}

	type Text struct {
		Type string `xml:"type,attr"`
		Body string `xml:",chardata"`
	}

	type Entry struct {
		Title     string  `xml:"title"`
		ID        string  `xml:"id"`
		Link      []Link  `xml:"link"`
		Published TimeStr `xml:"published"`
		Updated   TimeStr `xml:"updated"`
		Author    *Person `xml:"author"`
		Summary   *Text   `xml:"summary"`
		Content   *Text   `xml:"content"`
	}

	type Feed struct {
		XMLName xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
		Title   string   `xml:"title"`
		ID      string   `xml:"id"`
		Link    []Link   `xml:"link"`
		Updated TimeStr  `xml:"updated"`
		Author  *Person  `xml:"author"`
		Entry   []*Entry `xml:"entry"`
	}

	feed := Feed{
		Title: "The Sourcegraph Blog",
		ID:    conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.BlogIndex)).String(),
		Entry: make([]*Entry, len(posts)),
	}

	if len(posts) > 0 {
		feed.Updated = Time(posts[0].Date)
	}

	// Pagination links for the atom feed, for information on these see:
	//
	// http://tools.ietf.org/html/rfc5005#section-3
	// http://stackoverflow.com/questions/1301392/pagination-in-feeds-like-atom-and-rss
	//
	u := conf.AppURL(ctx).ResolveReference(r.URL)
	feed.Link = append(feed.Link, Link{
		Rel:  "self",
		Href: u.String(),
		Type: "application/atom+xml",
	})

	// Next page link:
	next := *u
	v := next.Query()
	v.Set("Page", strconv.Itoa(opt.Page+1))
	next.RawQuery = v.Encode()
	feed.Link = append(feed.Link, Link{
		Rel:  "next",
		Href: next.String(),
		Type: "application/atom+xml",
	})

	// Previous page link:
	prev := *u
	v = prev.Query()
	v.Set("Page", strconv.Itoa(opt.Page-1))
	prev.RawQuery = v.Encode()
	feed.Link = append(feed.Link, Link{
		Rel:  "prev",
		Href: prev.String(),
		Type: "application/atom+xml",
	})

	for i, p := range posts {
		post := resp.Posts[i]
		absURL := conf.AppURL(ctx).ResolveReference(router.Rel.URLToBlogPost(p.URL))

		body, err := htmlutil.MakeURLsAbsolute(string(post.Body), absURL)
		if err != nil {
			return err
		}

		entry := &Entry{
			Title: p.Title,
			ID:    absURL.String(),
			Link: []Link{
				{Href: absURL.String(), Rel: "alternate", Type: "text/html"},
			},
			Updated: Time(p.Date),
			Author: &Person{
				Name: fmt.Sprintf("%s (%s)", p.Author.Name, p.Author.Login),
			},
			Content: &Text{
				Type: "html",
				Body: string(sanitizeHTML(body)),
			},
		}
		if len(p.Author.Login) > 0 {
			entry.Author.URI = conf.AppURL(ctx).ResolveReference(router.Rel.URLToUser(p.Author.Login)).String()
		}
		feed.Entry[i] = entry
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Cache-Control", "max-age=900")
	w.Header().Set("Content-Type", "application/atom+xml; charset=utf-8")
	err = xml.NewEncoder(w).Encode(&feed)
	if err != nil {
		return err
	}
	return nil
}

// blogResolveSlug takes a slug string, which can either be of any legacy style:
//
// A old slug for redirection (see blogSlugRedirects map).
//
// A old slug for resolution into an ID (see blogSlugAliases map).
//
// A new Tumblr URL in the form of 1231231/some-nice-name.
//
// Regardless, it always returns the Tumblr post ID string.
func blogResolveSlug(slug string) (string, error) {
	// Check for slugs that used to be redirects, we treat them as an alias here
	// and resolve them below into an ID (see blogSlugAliases).
	newSlug, ok := blogSlugRedirects[slug]
	if ok {
		slug = newSlug
	}

	// Check for slug aliases.
	newSlug, ok = blogSlugAliases[slug]
	if ok {
		slug = newSlug
	}

	// Check if we have a Tumblr URL, where the first component is the post ID.
	if split := strings.Split(slug, "/"); len(split) >= 2 {
		id := split[0]
		slug = id
	}

	// Validate that the post ID is an integer.
	_, err := strconv.ParseInt(slug, 10, 64)
	if err != nil {
		return "", &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("no blog post found with slug %q (expected ID)", slug),
		}
	}
	return slug, nil
}

func serveBlogPost(w http.ResponseWriter, r *http.Request) error {
	ctx := httpctx.FromRequest(r)

	// Resolve the slug into a Tumblr post ID.
	slug := mux.Vars(r)["Slug"]
	id, err := blogResolveSlug(slug)
	if err != nil {
		return err
	}

	// Fetch the blog post.
	postsResp, err := tumblrBlog.Posts(PostsOpts{
		ID:              id,
		DisableSanitize: true,
	})
	if err != nil {
		return err
	}

	// Validate that we only got one post.
	if len(postsResp.Posts) != 1 {
		return &errcode.HTTPErr{
			Status: http.StatusNotFound,
			Err:    fmt.Errorf("expected one blog post; found %d", len(postsResp.Posts)),
		}
	}
	post := postsResp.Posts[0]
	meta := newBlogPostMeta(post)

	// All old URLs redirect to their new Tumblr URL.
	thisURL := conf.AppURL(ctx).ResolveReference(router.Rel.URLToBlogPost(slug)).String()
	canonicalURL := conf.AppURL(ctx).ResolveReference(router.Rel.URLToBlogPost(meta.URL))
	if thisURL != canonicalURL.String() {
		http.Redirect(w, r, canonicalURL.String(), http.StatusMovedPermanently)
		return nil
	}

	// Note: post.Body is valid HTML already because it is preprocessed and
	// sanitized in tumblr.Posts; see tumblr.go for more details.
	html := template.HTML(post.Body)

	return tmpl.Exec(r, w, "blog/post.html", http.StatusOK, nil, &struct {
		Meta        *blogPostMeta
		HTML        template.HTML
		URL         *url.URL
		RobotsIndex bool
		tmpl.Common
	}{
		Meta:        meta,
		HTML:        html,
		URL:         conf.AppURL(ctx).ResolveReference(canonicalURL),
		RobotsIndex: true,
	})
}
