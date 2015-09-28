package app

import (
	"math"
	"reflect"
	"strconv"

	"github.com/google/go-querystring/query"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// paginatePrevNext returns a list of page links that contain URLs and text
// labels for navigating to previous and next pages in a paged result
// set. currentSchema should be an *Options struct (e.g.,
// sourcegraph.RepoListOptions) with an embedded client.ListOptions
// struct. Those ListOptions along with listResponse are used to
// generate the page links.
func paginatePrevNext(currentSchema interface{}, listResponse sourcegraph.ListResponse) ([]pageLink, error) {
	return (&paginationPrevNext{CurrentSchema: currentSchema, listResponse: listResponse}).PageLinks(), nil
}

type paginationPrevNext struct {
	CurrentSchema interface{}
	listResponse  sourcegraph.ListResponse
}

func (p *paginationPrevNext) PageLinks() []pageLink {
	currentPage := listOptions(p.CurrentSchema).PageOrDefault()

	var links []pageLink

	prev := pageLink{
		Label:    prevPageDef,
		Disabled: currentPage == 1,
	}
	if !prev.Disabled {
		prev.URL = queryForPage(p.CurrentSchema, currentPage-1)
	}
	links = append(links, prev)

	next := pageLink{
		Label:    nextPageDef,
		Disabled: !p.listResponse.HasMore,
	}
	if !next.Disabled {
		next.URL = queryForPage(p.CurrentSchema, currentPage+1)
	}
	links = append(links, next)

	return links
}

// paginate returns a list of page links that contain URLs and text
// labels for referencing all of the pages in a paged result
// set. currentSchema should be an *Options struct (e.g.,
// sourcegraph.RepoListOptions) with an embedded client.ListOptions
// struct. Those ListOptions along with totalItems are used to
// generate the page links.
func paginate(currentSchema interface{}, totalItems int) ([]pageLink, error) {
	return (&pagination{currentSchema, totalItems}).PageLinks(), nil
}

type pagination struct {
	CurrentSchema interface{}
	totalItems    int
}

type pageLink struct {
	URL      string
	Label    string
	Current  bool
	Disabled bool
}

const (
	prevPageDef   = "\u25C0" // left triangle
	nextPageDef   = "\u25B6" // right triangle
	elidedPageDef = "\u2026" // ellipsis
	maxPageLinks  = 8
)

func (p *pagination) PageLinks() []pageLink {
	listOpts := listOptions(p.CurrentSchema)
	currentPage, perPage := listOpts.PageOrDefault(), listOpts.PerPageOrDefault()
	numPages := int(math.Ceil(float64(p.totalItems) / float64(perPage)))

	if numPages <= 1 {
		return nil
	}

	var links []pageLink

	prev := pageLink{
		Label:    prevPageDef,
		Disabled: currentPage == 1,
	}
	if !prev.Disabled {
		prev.URL = queryForPage(p.CurrentSchema, currentPage-1)
	}
	links = append(links, prev)

	// Numbered page links.
	for page := 1; page <= numPages; page++ {
		if numPages > maxPageLinks && (page > 2 && page < numPages-1) {
			d := abs(currentPage - page)
			if d > maxPageLinks/2 {
				continue
			} else if d == maxPageLinks/2 {
				links = append(links, pageLink{Label: elidedPageDef, Disabled: true})
				continue
			}
		}
		links = append(links, pageLink{
			URL:     queryForPage(p.CurrentSchema, page),
			Label:   strconv.Itoa(page),
			Current: currentPage == page,
		})
	}

	next := pageLink{
		Label:    nextPageDef,
		Disabled: currentPage >= numPages,
	}
	if !next.Disabled {
		next.URL = queryForPage(p.CurrentSchema, currentPage+1)
	}
	links = append(links, next)

	return links
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func listOptions(currentSchema interface{}) sourcegraph.ListOptions {
	var listOpts sourcegraph.ListOptions
	st := reflect.ValueOf(currentSchema)
	if v := st.FieldByName("ListOptions"); v.IsValid() {
		listOpts = v.Interface().(sourcegraph.ListOptions)
	} else {
		panic("pagination.CurrentSchema has no ListOptions field: " + st.String())
	}
	return listOpts
}

func queryForPage(currentSchema interface{}, page int) string {
	qs, err := query.Values(currentSchema)
	if err != nil {
		panic("queryForPage: " + err.Error())
	}
	if page > 1 {
		qs.Set("Page", strconv.Itoa(page))
	} else {
		delete(qs, "Page")
	}
	s := qs.Encode()
	return "?" + s
}
