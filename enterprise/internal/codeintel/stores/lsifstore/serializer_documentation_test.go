package lsifstore

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/protocol"
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
			want: autogold.Want("basic", nil),
		},
		{
			page: &precise.DocumentationPageData{
				Tree: &precise.DocumentationNode{
					PathID: "/",
					Children: nil,
				},
			},
			want: autogold.Want("nil children would be encoded as JSON array not null", nil),
		},
		{
			page: &precise.DocumentationPageData{
				Tree: &precise.DocumentationNode{
					PathID: "/",
					Documentation: protocol.Documentation{},
				},
			},
			want: autogold.Want("nil Documentation.Tags would be encoded as JSON array not null", nil),
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
