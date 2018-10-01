package phabricator

import (
	"fmt"
	"testing"
)

// TODO: Is it possible to pull this from `conf.Get().Phabricator`?
// The commented out code below tries this but it says its not set.
const phabURL = "https://phabricator.sgdev.org"
const apiToken = "api-5ngcoazuyugsoiuumxxuhd4g26ac"

// var apiToken string
//
// func init() {
// phabs := conf.Get().Phabricator
// fmt.Println(phabs)
// for _, phab := range phabs {
// if phab.Url == phabURL {
// apiToken = phab.Token
// }
// }
// }

var diff122 = `diff --git a/math.go b/math.go
--- a/math.go
+++ b/math.go
@@ -15,3 +15,7 @@
 func Multiply(a, b int) int {
 	return a * b
 }
+
+func SomeNewFunction() {
+	// This function will eventually do stuff
+}
diff --git a/.arcconfig b/.arcconfig
--- a/.arcconfig
+++ b/.arcconfig
@@ -1,3 +1,3 @@
 {
-    "phabricator.uri" : "http://phabricator.sgdev.org/"
+    "phabricator.uri" : "https://phabricator.sgdev.org/"
 }

`

// This is probably a flakey test because it's relying on phabricator results
func TestGetRawDiff(t *testing.T) {
	t.Skip("TODO(isaac): https://github.com/sourcegraph/sourcegraph/issues/12724")
	if apiToken == "" {
		// Should we cause this test to fail if not configured?
		t.Error("no api token provided")
		return
	}

	client := NewClient(phabURL, apiToken)

	diff, err := client.GetRawDiff(122)
	if err != nil {
		t.Fatalf("unexpected error getting raw diff: %v", err)
	}

	if string(diff) != diff122 {
		fmt.Println("0'", string(diff), "'")
		fmt.Println("1'", diff122, "'")
		t.Errorf("result doesn't match diff")
	}
}
