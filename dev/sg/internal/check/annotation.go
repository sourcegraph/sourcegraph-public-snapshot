package check

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

func generateAnnotation(category string, check string, content string) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return // do nothing
	}

	// set up annotations dir
	annotationsDir := filepath.Join(repoRoot, "annotations")
	os.MkdirAll(annotationsDir, os.ModePerm)

	// write annotation
	path := filepath.Join(annotationsDir, fmt.Sprintf("%s: %s.md", category, check))
	_ = os.WriteFile(path, []byte(content+"\n"), os.ModePerm)
}
