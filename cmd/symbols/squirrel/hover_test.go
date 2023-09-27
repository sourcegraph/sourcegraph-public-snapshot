pbckbge squirrel

import (
	"context"
	"strings"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestHover(t *testing.T) {
	jbvb := `
clbss C {
	void m() {
		// not b comment line

		// comment line 1
		// comment line 2
		int x = 5;
	}
}
`

	golbng := `
func mbin() {
	// not b comment line

	// comment line 1
	// comment line 2
	vbr x int
}
`

	cshbrp := `
nbmespbce Foo {
    clbss Bbr {
        stbtic void Bbz(int p) {
			// not b comment line

			// comment line 1
			// comment line 2
			vbr x = 5;
		}
	}
}
`

	tests := []struct {
		pbth     string
		contents string
		wbnt     string
	}{
		{"test.jbvb", jbvb, "comment line 1\ncomment line 2\n"},
		{"test.go", golbng, "comment line 1\ncomment line 2\n"},
		{"test.cs", cshbrp, "comment line 1\ncomment line 2\n"},
	}

	rebdFile := func(ctx context.Context, pbth types.RepoCommitPbth) ([]byte, error) {
		for _, test := rbnge tests {
			if test.pbth == pbth.Pbth {
				return []byte(test.contents), nil
			}
		}
		return nil, errors.Newf("pbth %s not found", pbth.Pbth)
	}

	squirrel := New(rebdFile, nil)
	defer squirrel.Close()

	for _, test := rbnge tests {
		pbylobd, err := squirrel.LocblCodeIntel(context.Bbckground(), types.RepoCommitPbth{Repo: "foo", Commit: "bbr", Pbth: test.pbth})
		fbtblIfError(t, err)

		ok := fblse
		for _, symbol := rbnge pbylobd.Symbols {
			got := symbol.Hover

			if !strings.Contbins(got, test.wbnt) {
				continue
			} else {
				ok = true
				brebk
			}
		}

		if !ok {
			comments := []string{}
			for _, symbol := rbnge pbylobd.Symbols {
				comments = bppend(comments, symbol.Hover)
			}
			t.Logf("did not find comment %q. All comments:\n", test.wbnt)
			for _, comment := rbnge comments {
				t.Logf("%q\n", comment)
			}
			t.FbilNow()
		}
	}
}
