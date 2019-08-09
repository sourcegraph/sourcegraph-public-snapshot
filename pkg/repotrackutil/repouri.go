package repotrackutil

import (
	"regexp"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

var trackedRepo = []string{
	"github.com/kubernetes/kubernetes",
	"github.com/gorilla/mux",
	"github.com/golang/go",
	"sourcegraph/sourcegraph",
}
var trackedRepoRe = regexp.MustCompile(`\b(` + strings.Join(trackedRepo, "|") + `)\b`)

// GetTrackedRepo guesses which repo a request URL path is for. It only looks
// at a certain subset of repos for its guess.
func GetTrackedRepo(repoPath api.RepoName) string {
	m := trackedRepoRe.FindStringSubmatch(string(repoPath))
	if len(m) == 0 {
		return "unknown"
	}
	return m[1]
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_877(size int) error {
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
