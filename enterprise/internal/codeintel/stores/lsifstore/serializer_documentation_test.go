package lsifstore

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

// Note: You can `go test ./pkg -update` to update the expected `want` values in these tests.
// See https://github.com/hexops/autogold for more information.

func TestSerializeDocumentationPageData(t *testing.T) {
	testCases := []struct {
		page *precise.DocumentationPageData
		want autogold.Value
	}{
		{
			page: &precise.DocumentationPageData{
				Tree: &precise.DocumentationNode{
					PathID: "/",
					Children: []precise.DocumentationNodeChild{
						{PathID: "/somelinkedpage"},
						{Node: &precise.DocumentationNode{
							PathID: "/#main",
						}},
						{PathID: "/somelinkedpage2"},
						{Node: &precise.DocumentationNode{
							PathID: "/subpkg",
							Children: []precise.DocumentationNodeChild{
								{Node: &precise.DocumentationNode{
									PathID: "/subpkg#Router",
								}},
							},
						}},
					},
				},
			},
			want: autogold.Want("basic", &precise.DocumentationPageData{Tree: &precise.DocumentationNode{
				PathID:        "/",
				Documentation: protocol.Documentation{Tags: []protocol.Tag{}},
				Children: []precise.DocumentationNodeChild{
					{PathID: "/somelinkedpage"},
					{Node: &precise.DocumentationNode{
						PathID: "/#main",
						Documentation: protocol.Documentation{
							Tags: []protocol.Tag{},
						},
						Children: []precise.DocumentationNodeChild{},
					}},
					{PathID: "/somelinkedpage2"},
					{Node: &precise.DocumentationNode{
						PathID:        "/subpkg",
						Documentation: protocol.Documentation{Tags: []protocol.Tag{}},
						Children: []precise.DocumentationNodeChild{{Node: &precise.DocumentationNode{
							PathID:        "/subpkg#Router",
							Documentation: protocol.Documentation{Tags: []protocol.Tag{}},
							Children:      []precise.DocumentationNodeChild{},
						}}},
					}},
				},
			}}),
		},
		{
			page: &precise.DocumentationPageData{
				Tree: &precise.DocumentationNode{
					PathID:   "/",
					Children: nil,
				},
			},
			want: autogold.Want("nil children would be encoded as JSON array not null", &precise.DocumentationPageData{Tree: &precise.DocumentationNode{
				PathID:        "/",
				Documentation: protocol.Documentation{Tags: []protocol.Tag{}},
				Children:      []precise.DocumentationNodeChild{},
			}}),
		},
		{
			page: &precise.DocumentationPageData{
				Tree: &precise.DocumentationNode{
					PathID:        "/",
					Documentation: protocol.Documentation{},
				},
			},
			want: autogold.Want("nil Documentation.Tags would be encoded as JSON array not null", &precise.DocumentationPageData{Tree: &precise.DocumentationNode{
				PathID:        "/",
				Documentation: protocol.Documentation{Tags: []protocol.Tag{}},
				Children:      []precise.DocumentationNodeChild{},
			}}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			serializer := NewSerializer()
			encoded, err := serializer.MarshalDocumentationPageData(tc.page)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := serializer.UnmarshalDocumentationPageData(encoded)
			if err != nil {
				t.Fatal(err)
			}
			got := decoded
			tc.want.Equal(t, got)
		})
	}
}

func TestSerializeDocumentationPathInfoData(t *testing.T) {
	testCases := []struct {
		pathInfo *precise.DocumentationPathInfoData
		want     autogold.Value
	}{
		{
			pathInfo: &precise.DocumentationPathInfoData{
				PathID:   "/",
				IsIndex:  true,
				Children: []string{"/sub", "/sub/pkg", "/sub/pkg/sub/pkg"},
			},
			want: autogold.Want("basic", &precise.DocumentationPathInfoData{
				PathID: "/", IsIndex: true,
				Children: []string{
					"/sub",
					"/sub/pkg",
					"/sub/pkg/sub/pkg",
				},
			}),
		},
		{
			pathInfo: &precise.DocumentationPathInfoData{
				PathID:   "/",
				IsIndex:  true,
				Children: nil,
			},
			want: autogold.Want("nil children would be encoded as JSON array not null", &precise.DocumentationPathInfoData{
				PathID: "/", IsIndex: true,
				Children: []string{},
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			serializer := NewSerializer()
			encoded, err := serializer.MarshalDocumentationPathInfoData(tc.pathInfo)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := serializer.UnmarshalDocumentationPathInfoData(encoded)
			if err != nil {
				t.Fatal(err)
			}
			got := decoded
			tc.want.Equal(t, got)
		})
	}
}
