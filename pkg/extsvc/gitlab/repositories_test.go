package gitlab

import (
	"context"
	"reflect"
	"testing"
)

func TestListTree(t *testing.T) {
	mock := mockHTTPResponseBody{
		responseBody: `
[
  {
    "id": "a1e8f8d745cc87e3a9248358d9352bb7f9a0aeba",
    "name": "html",
    "type": "tree",
    "path": "files/html",
    "mode": "040000"
  },
  {
    "id": "4535904260b1082e14f867f7a24fd8c21495bde3",
    "name": "images",
    "type": "tree",
    "path": "files/images",
    "mode": "040000"
  }
]
`}
	c := newTestClient(t)
	c.httpClient.Transport = &mock

	want := []*Tree{
		{
			ID:   "a1e8f8d745cc87e3a9248358d9352bb7f9a0aeba",
			Name: "html",
			Type: "tree",
			Path: "files/html",
			Mode: "040000",
		},
		{
			ID:   "4535904260b1082e14f867f7a24fd8c21495bde3",
			Name: "images",
			Type: "tree",
			Path: "files/images",
			Mode: "040000",
		},
	}

	// Test first fetch (cache empty)
	tree, err := c.ListTree(context.Background(), ListTreeOp{ProjPathWithNamespace: "n1/n2/r"})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Error("tree == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to miss cache once", mock.count)
	}
	if !reflect.DeepEqual(tree, want) {
		t.Errorf("got tree %+v, want %+v", tree, &want)
	}

	// Note: since caching is not currently implemented for this endpoint, we don't test it

	// Test the `NoCache: true` option
	tree, err = c.ListTree(context.Background(), ListTreeOp{ProjPathWithNamespace: "n1/n2/r", CommonOp: CommonOp{NoCache: true}})
	if err != nil {
		t.Fatal(err)
	}
	if tree == nil {
		t.Error("tree == nil")
	}
	if mock.count != 2 {
		t.Errorf("mock.count == %d, expected to miss cache once", mock.count)
	}
	if !reflect.DeepEqual(tree, want) {
		t.Errorf("got tree %+v, want %+v", tree, &want)
	}

}
