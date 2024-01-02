package dotcom

import (
	"os"
	"strings"
	"testing"

	"github.com/bazelbuild/rules_go/go/runfiles"
	"github.com/google/go-cmp/cmp"
)

func TestGeneratedFileMatchesFoo(t *testing.T) {
	generatedFile := strings.TrimSpace(os.Getenv("GENERATED_FILE"))
	goldenFile := strings.TrimSpace(os.Getenv("GOLDEN_FILE"))
	if generatedFile == "" {
		t.Fatal("Need GENERATED_FILE in env")
	}
	if goldenFile == "" {
		t.Fatal("Need GOLDEN_FILE")
	}
	generatedFile, err := runfiles.Rlocation(generatedFile)
	if err != nil {
		t.Fatalf("Cannot resolve location of GENERATED_FILE: %s", err)
	}
	goldenFile, err = runfiles.Rlocation(goldenFile)
	if err != nil {
		t.Fatalf("Cannot resolve location of GOLDEN_FILE: %s", err)
	}
	data, err := os.ReadFile(generatedFile)
	if err != nil {
		t.Error(os.Getenv("PWD"))
		t.Fatalf("Reading generated file %s failed: %s", generatedFile, err)
	}
	generatedContent := string(data)
	data, err = os.ReadFile(goldenFile)
	if err != nil {
		t.Fatalf("Reading golden file %s failed: %s", goldenFile, err)
	}
	goldenContent := string(data)
	if diff := cmp.Diff(goldenContent, generatedContent); diff != "" {
		t.Errorf("%s different from %s:\n%s\n", generatedFile, goldenFile, diff)
		t.Fatal("Please run bazel run cmd/cody-gateway/internal/httpapi/attribution/dotcom:write_genql_yaml")
	}
}
