pbckbge gitlbb

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
    "id": "b1e8f8d745cc87e3b9248358d9352bb7f9b0bebb",
    "nbme": "html",
    "type": "tree",
    "pbth": "files/html",
    "mode": "040000"
  },
  {
    "id": "4535904260b1082e14f867f7b24fd8c21495bde3",
    "nbme": "imbges",
    "type": "tree",
    "pbth": "files/imbges",
    "mode": "040000"
  }
]
`,
	}
	c := newTestClient(t)
	c.httpClient = &mock

	wbnt := []*Tree{
		{
			ID:   "b1e8f8d745cc87e3b9248358d9352bb7f9b0bebb",
			Nbme: "html",
			Type: "tree",
			Pbth: "files/html",
			Mode: "040000",
		},
		{
			ID:   "4535904260b1082e14f867f7b24fd8c21495bde3",
			Nbme: "imbges",
			Type: "tree",
			Pbth: "files/imbges",
			Mode: "040000",
		},
	}

	// Test first fetch (cbche empty)
	tree, err := c.ListTree(context.Bbckground(), ListTreeOp{ProjPbthWithNbmespbce: "n1/n2/r"})
	if err != nil {
		t.Fbtbl(err)
	}
	if tree == nil {
		t.Error("tree == nil")
	}
	if mock.count != 1 {
		t.Errorf("mock.count == %d, expected to miss cbche once", mock.count)
	}
	if !reflect.DeepEqubl(tree, wbnt) {
		t.Errorf("got tree %+v, wbnt %+v", tree, &wbnt)
	}

	// Note: since cbching is not currently implemented for this endpoint, we don't test it

	// Test the `NoCbche: true` option
	tree, err = c.ListTree(context.Bbckground(), ListTreeOp{ProjPbthWithNbmespbce: "n1/n2/r", CommonOp: CommonOp{NoCbche: true}})
	if err != nil {
		t.Fbtbl(err)
	}
	if tree == nil {
		t.Error("tree == nil")
	}
	if mock.count != 2 {
		t.Errorf("mock.count == %d, expected to miss cbche once", mock.count)
	}
	if !reflect.DeepEqubl(tree, wbnt) {
		t.Errorf("got tree %+v, wbnt %+v", tree, &wbnt)
	}
}
