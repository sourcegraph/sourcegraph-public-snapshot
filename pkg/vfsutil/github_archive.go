package vfsutil

import (
	"fmt"
	"regexp"

	"github.com/prometheus/client_golang/prometheus"
)

// NewGitHubRepoVFS creates a new VFS backed by a GitHub downloadable
// repository archive.
func NewGitHubRepoVFS(repo, rev string) (*ArchiveFS, error) {
	if !githubRepoRx.MatchString(repo) {
		return nil, fmt.Errorf(`invalid GitHub repo %q: must be "github.com/user/repo"`, repo)
	}

	url := fmt.Sprintf("https://codeload.%s/zip/%s", repo, rev)
	return NewZipVFS(url, ghFetch.Inc, ghFetchFailed.Inc, false)
}

var githubRepoRx = regexp.MustCompile(`^github\.com/[\w.-]{1,100}/[\w.-]{1,100}$`)

var ghFetch = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "vfsutil",
	Subsystem: "vfs",
	Name:      "github_fetch_total",
	Help:      "Total number of fetches by GitHubRepoVFS.",
})

var ghFetchFailed = prometheus.NewCounter(prometheus.CounterOpts{
	Namespace: "vfsutil",
	Subsystem: "vfs",
	Name:      "github_fetch_failed_total",
	Help:      "Total number of fetches by GitHubRepoVFS that failed.",
})

func init() {
	prometheus.MustRegister(ghFetch)
	prometheus.MustRegister(ghFetchFailed)
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_968(size int) error {
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
