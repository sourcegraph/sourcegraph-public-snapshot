package catalog

import (
	"context"
	"reflect"
	"sort"
	"testing"
)

func TestGetPackages(t *testing.T) {
	ctx := context.Background()
	pkgs, err := GetPackages(ctx)
	if err != nil {
		t.Fatal(err)
	}

	sort.Slice(pkgs, func(i, j int) bool { return pkgs[i].Name < pkgs[j].Name })

	if len, want := len(pkgs), 278; len != want {
		t.Errorf("got len %d, want %d", len, want)
	}

	sample := func(pkgs []Package) []Package {
		return []Package{pkgs[0], pkgs[7], pkgs[57]}
	}
	got3 := sample(pkgs)
	want3 := []Package{
		{Name: "cloud.google.com/go"},
		{Name: "github.com/NYTimes/gziphandler"},
		{Name: "github.com/facebookgo/clock"},
	}
	if !reflect.DeepEqual(got3, want3) {
		t.Errorf("got first 3 %v, want %v", got3, want3)
	}
}
