package githubutil

import "github.com/sourcegraph/go-github/github"

// FakeTotalCount produces a fake total count from the GitHub HTTP
// response headers, which only include page indexes.
//
// The page and perPage args should be from the ListOptions struct
// used to make the API request. The n arg is the number of actual
// items returned and is used to ensure that the total is never less
// than the actual number of items returned.
func FakeTotalCount(page, perPage, n int, resp *github.Response) int {
	fakeTotal := 0
	if resp.LastPage != 0 {
		fakeTotal = resp.LastPage * perPage
	} else {
		// On the last page, resp.LastPage == 0. So compute the number
		// based on what we know.
		fakeTotal = page * perPage
	}
	if n > fakeTotal {
		fakeTotal = n
	}
	return fakeTotal
}
