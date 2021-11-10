package queryrunner

import (
	"context"
	"testing"
)

func TestAThing(t *testing.T) {
	ctx := context.Background()
	results, err := computeSearch(ctx, "repo:^github\\.com/sourcegraph/sourcegraph$ file:go\\.mod$ go\\s*(\\d\\.\\d+)")
	if err != nil {
		t.Fatal(err)
	}

	for _, computeResult := range results {
		t.Logf("result: repo: %v commit:%v", computeResult.RepoName(), computeResult.Revhash())
		for val, count := range computeResult.Counts() {
			t.Logf("match value: %s count: %d", val, count)
		}
	}
}
