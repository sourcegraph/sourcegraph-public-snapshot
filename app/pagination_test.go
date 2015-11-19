package app

import (
	"strings"
	"testing"

	"github.com/kr/pretty"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

func TestPagination_PageLinks(t *testing.T) {
	type myOptions struct {
		A string `url:",omitempty"`
		sourcegraph.ListOptions
	}

	must := func(pg []pageLink, err error) []pageLink {
		if err != nil {
			t.Error(err)
		}
		return pg
	}

	tests := map[string]struct {
		got  []pageLink
		want []pageLink
	}{
		"single page": {
			got: must(paginatePrevNext(myOptions{}, sourcegraph.StreamResponse{HasMore: false})),
			want: []pageLink{
				{URL: "", Label: prevPageDef, Disabled: true},
				{Label: nextPageDef, Disabled: true},
			},
		},
		"first page": {
			got: must(paginatePrevNext(myOptions{}, sourcegraph.StreamResponse{HasMore: true})),
			want: []pageLink{
				{URL: "", Label: prevPageDef, Disabled: true},
				{URL: "?Page=2", Label: nextPageDef},
			},
		},
		"middle page": {
			got: must(paginatePrevNext(myOptions{ListOptions: sourcegraph.ListOptions{Page: 5}}, sourcegraph.StreamResponse{HasMore: true})),
			want: []pageLink{
				{URL: "?Page=4", Label: prevPageDef},
				{URL: "?Page=6", Label: nextPageDef},
			},
		},
		"last page": {
			got: must(paginatePrevNext(myOptions{ListOptions: sourcegraph.ListOptions{Page: 10}}, sourcegraph.StreamResponse{HasMore: false})),
			want: []pageLink{
				{URL: "?Page=9", Label: prevPageDef},
				{URL: "", Label: nextPageDef, Disabled: true},
			},
		},
		"after last page": {
			got: must(paginatePrevNext(myOptions{ListOptions: sourcegraph.ListOptions{Page: 100}}, sourcegraph.StreamResponse{HasMore: false})),
			want: []pageLink{
				{URL: "?Page=99", Label: prevPageDef},
				{URL: "", Label: nextPageDef, Disabled: true},
			},
		},

		"no pages": {
			got:  must(paginate(myOptions{}, 0)),
			want: nil,
		},
		"merge with other opts": {
			got: must(paginate(myOptions{A: "a"}, 11)),
			want: []pageLink{
				{URL: "", Label: prevPageDef, Disabled: true},
				{URL: "?A=a", Label: "1", Current: true},
				{URL: "?A=a&Page=2", Label: "2"},
				{URL: "?A=a&Page=2", Label: nextPageDef},
			},
		},
		"many pages": {
			got: must(paginate(myOptions{A: "a", ListOptions: sourcegraph.ListOptions{Page: 2}}, 25)),
			want: []pageLink{
				{URL: "?A=a", Label: prevPageDef},
				{URL: "?A=a", Label: "1"},
				{URL: "?A=a&Page=2", Label: "2", Current: true},
				{URL: "?A=a&Page=3", Label: "3"},
				{URL: "?A=a&Page=3", Label: nextPageDef},
			},
		},
		"limit total number of pages shown": {
			got: must(paginate(myOptions{ListOptions: sourcegraph.ListOptions{Page: 57}}, 2500)),
			want: []pageLink{
				{URL: "?Page=56", Label: prevPageDef},
				{URL: "?", Label: "1"},
				{URL: "?Page=2", Label: "2"},
				{Label: elidedPageDef, Disabled: true},
				{URL: "?Page=54", Label: "54"},
				{URL: "?Page=55", Label: "55"},
				{URL: "?Page=56", Label: "56"},
				{URL: "?Page=57", Label: "57", Current: true},
				{URL: "?Page=58", Label: "58"},
				{URL: "?Page=59", Label: "59"},
				{URL: "?Page=60", Label: "60"},
				{Label: elidedPageDef, Disabled: true},
				{URL: "?Page=249", Label: "249"},
				{URL: "?Page=250", Label: "250"},
				{URL: "?Page=58", Label: nextPageDef},
			},
		},
	}
	for label, test := range tests {
		if diff := pretty.Diff(test.got, test.want); len(diff) > 0 {
			t.Errorf("%s: links did not match expected\n%s", label, strings.Join(diff, "\n"))
			continue
		}
	}
}
