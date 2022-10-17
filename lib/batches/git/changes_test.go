package git

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestChangesInDiff(t *testing.T) {
	const input = `diff --git README.md README.md
index c9644dd..2552420 100644
--- README.md
+++ README.md
@@ -1,2 +1,3 @@
 # Welcome to the README
 foobar
+barfoo and what else?
diff --git a_new_file_appears.txt a_new_file_appears.txt
new file mode 100644
index 0000000..09f946b
--- /dev/null
+++ a_new_file_appears.txt
@@ -0,0 +1 @@
+boom! like magic it appears
diff --git another_file.txt another_cool_file.txt
similarity index 100%
rename from another_file.txt
rename to another_cool_file.txt
diff --git yet_another_file.txt yet_another_file.txt
deleted file mode 100644
index c27a40c..0000000
--- yet_another_file.txt
+++ /dev/null
@@ -1,3 +0,0 @@
-this is yet another file
-this time though
-with tree lines
`

	changes, err := ChangesInDiff([]byte(input))
	if err != nil {
		t.Fatal(err)
	}

	want := Changes{
		Modified: []string{"README.md"},
		Added:    []string{"a_new_file_appears.txt"},
		Deleted:  []string{"yet_another_file.txt"},
		Renamed:  []string{"another_cool_file.txt"},
	}

	if !cmp.Equal(want, changes) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, changes))
	}
}
