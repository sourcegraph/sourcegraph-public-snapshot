package repotrackutil

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func TestGetTrackedRepo(t *testing.T) {
	cases := []struct {
		Path        api.RepoName
		TrackedRepo string
	}{
		// Top-level view
		{"/github.com/kubernetes/kubernetes", "github.com/kubernetes/kubernetes"},
		// Code view
		{"/github.com/kubernetes/kubernetes@master/-/tree/README.md", "github.com/kubernetes/kubernetes"},

		// Unrelated repo
		{"/github.com/gorilla/muxy@master/-/tree/mux.go", "unknown"},
		{"/github.com/gorilla/muxy", "unknown"},

		// Unrelated URL
		{"/blog/133554180524/announcing-the-sourcegraph-developer-release-the", "unknown"},

		// Corner case
		{"", "unknown"}, {"/", "unknown"},
	}
	for _, c := range cases {
		got := GetTrackedRepo(c.Path)
		if got != c.TrackedRepo {
			t.Errorf("GetTrackedRepo(%#v) == %#v != %#v", c.Path, got, c.TrackedRepo)
		}
	}
	// a trackedRepo must always be tracked
	for _, r := range trackedRepo {
		if GetTrackedRepo(api.RepoName(r)) != string(r) {
			t.Errorf("Repo should be tracked: %v", r)
		}
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_878(size int) error {
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
