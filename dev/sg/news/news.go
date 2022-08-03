package news

import (
	"context"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
)

func Render(ctx context.Context, build string) (string, error) {
	var md strings.Builder

	md.WriteString("# 📰 sg news\n")

	// Render some recent changes
	title, changes, err := repo.RecentSGChanges(build, repo.RecentChangesOpts{
		Count: 5,
		Color: false, // we render markdown
	})
	if err != nil {
		return "", err
	}
	md.WriteString("## " + title + "\n")
	for _, change := range changes {
		md.WriteString("- " + change + "\n")
	}

	// Render the final output
	return md.String(), nil
}
