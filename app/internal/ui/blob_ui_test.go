// +build uitest

package ui

import (
	"testing"
	"time"

	"golang.org/x/net/context"
)

func TestRenderReactComponent_Blob(t *testing.T) {
	ensureBundleJSFound(t)
	blob, stores, want := blobRenderTestdata(true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := renderReactComponent(ctx, "sourcegraph/blob/Blob", blob, &stores)
	if err != nil {
		t.Fatal(err)
	}

	if resp != want {
		t.Errorf("got\n%s\n\nwant\n%s", resp, want)
	}
}
