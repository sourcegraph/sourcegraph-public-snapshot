package query

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func prettyPrint(nodes []Node) string {
	var resultStr []string
	for _, node := range nodes {
		resultStr = append(resultStr, node.String())
	}
	return strings.Join(resultStr, " ")
}

func Test_LowercaseFieldNames(t *testing.T) {
	input := "rEpO:foo PATTERN"
	want := `(and "repo:foo" "PATTERN")`
	query, _ := ParseAndOr(input)
	got := prettyPrint(LowercaseFieldNames(query))
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatal(diff)
	}
}
