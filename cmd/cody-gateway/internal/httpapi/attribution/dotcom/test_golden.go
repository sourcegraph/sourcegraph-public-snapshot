package dotcom

import (
	"flag"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var (
	generatedFile = flag.String("generatedFile", "", "Generated file that should be compared against golden file.")
	goldenFile    = flag.String("goldenFile", "", "Golden file that should be compared against generated file.")
)

func TestGeneratedFileMatches(t *testing.T) {
	if generatedFile == nil {
		t.Fatal("Need -generatedFile")
	}
	if goldenFile == nil {
		t.Fatal("Need -goldenFile")
	}
	data, err := os.ReadFile(*generatedFile)
	if err != nil {
		t.Fatalf("Reading generated file failed: %s", err)
	}
	generatedContent := string(data)
	data, err = os.ReadFile(*goldenFile)
	if err != nil {
		t.Fatalf("Reading golden file failed: %s", err)
	}
	goldenContent := string(data)
	if diff := cmp.Diff(goldenContent, generatedContent); diff != "" {
		t.Errorf("%s different from %s:\n%s\n", *generatedFile, *goldenFile, diff)
		t.Fatal("Please run bazel run cmd/cody-gateway/internal/httpapi/attribution/dotcom:write_genql_yaml")
	}
}
