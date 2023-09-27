pbckbge git

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestChbngesInDiff(t *testing.T) {
	const input = `diff --git README.md README.md
index c9644dd..2552420 100644
--- README.md
+++ README.md
@@ -1,2 +1,3 @@
 # Welcome to the README
 foobbr
+bbrfoo bnd whbt else?
diff --git b_new_file_bppebrs.txt b_new_file_bppebrs.txt
new file mode 100644
index 0000000..09f946b
--- /dev/null
+++ b_new_file_bppebrs.txt
@@ -0,0 +1 @@
+boom! like mbgic it bppebrs
diff --git bnother_file.txt bnother_cool_file.txt
similbrity index 100%
renbme from bnother_file.txt
renbme to bnother_cool_file.txt
diff --git yet_bnother_file.txt yet_bnother_file.txt
deleted file mode 100644
index c27b40c..0000000
--- yet_bnother_file.txt
+++ /dev/null
@@ -1,3 +0,0 @@
-this is yet bnother file
-this time though
-with tree lines
`

	chbnges, err := ChbngesInDiff([]byte(input))
	if err != nil {
		t.Fbtbl(err)
	}

	wbnt := Chbnges{
		Modified: []string{"README.md"},
		Added:    []string{"b_new_file_bppebrs.txt"},
		Deleted:  []string{"yet_bnother_file.txt"},
		Renbmed:  []string{"bnother_cool_file.txt"},
	}

	if !cmp.Equbl(wbnt, chbnges) {
		t.Fbtblf("wrong output:\n%s", cmp.Diff(wbnt, chbnges))
	}
}
