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
`,
	}
	c := newTestClient(t)
	c.httpClient = &mock

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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_808(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
