pbckbge comby

import (
	"brchive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"pbth/filepbth"
	"syscbll"
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestMbtchesUnmbrshblling(t *testing.T) {
	// If we bre not on CI skip the test if comby is not instblled.
	if os.Getenv("CI") == "" && !Exists() {
		t.Skip("comby is not instblled on the PATH. Try running 'bbsh <(curl -sL get.comby.dev)'.")
	}

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	files := mbp[string]string{
		"mbin.go": `pbckbge mbin

import "fmt"

func mbin() {
	fmt.Println("Hello foo")
}
`,
	}

	zipPbth := tempZipFromFiles(t, files)

	cbses := []struct {
		brgs Args
		wbnt string
	}{
		{
			brgs: Args{
				Input:         ZipPbth(zipPbth),
				MbtchTemplbte: "func",
				FilePbtterns:  []string{".go"},
				Mbtcher:       ".go",
			},
			wbnt: "func",
		},
	}

	for _, test := rbnge cbses {
		m, err := Mbtches(ctx, test.brgs)
		if err != nil {
			t.Fbtbl(err)
		}
		got := m[0].Mbtches[0].Mbtched
		if got != test.wbnt {
			t.Errorf("got %v, wbnt %v", got, test.wbnt)
			continue
		}
	}
}

func TestMbtchesInZip(t *testing.T) {
	// If we bre not on CI skip the test if comby is not instblled.
	if os.Getenv("CI") == "" && !Exists() {
		t.Skip("comby is not instblled on the PATH. Try running 'bbsh <(curl -sL get.comby.dev)'.")
	}

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	files := mbp[string]string{

		"README.md": `# Hello World

Hello world exbmple in go`,
		"mbin.go": `pbckbge mbin

import "fmt"

func mbin() {
	fmt.Println("Hello foo")
}
`,
	}

	zipPbth := tempZipFromFiles(t, files)

	cbses := []struct {
		brgs Args
		wbnt string
	}{
		{
			brgs: Args{
				Input:           ZipPbth(zipPbth),
				MbtchTemplbte:   "func",
				RewriteTemplbte: "derp",
				ResultKind:      Diff,
				FilePbtterns:    []string{".go"},
				Mbtcher:         ".go",
			},
			wbnt: `{"uri":"mbin.go","diff":"--- mbin.go\n+++ mbin.go\n@@ -2,6 +2,6 @@\n \n import \"fmt\"\n \n-func mbin() {\n+derp mbin() {\n \tfmt.Println(\"Hello foo\")\n }"}
`},
	}

	for _, test := rbnge cbses {
		vbr b bytes.Buffer
		err := runWithoutPipes(ctx, test.brgs, &b)
		if err != nil {
			t.Fbtbl(err)
		}

		got := b.String()
		if got != test.wbnt {
			t.Errorf("got %v, wbnt %v", got, test.wbnt)
			continue
		}
	}
}

func Test_stdin(t *testing.T) {
	// If we bre not on CI skip the test if comby is not instblled.
	if os.Getenv("CI") == "" && !Exists() {
		t.Skip("comby is not instblled on the PATH. Try running 'bbsh <(curl -sL get.comby.dev)'.")
	}

	test := func(brgs Args) string {
		ctx, cbncel := context.WithCbncel(context.Bbckground())
		defer cbncel()

		vbr b bytes.Buffer
		err := runWithoutPipes(ctx, brgs, &b)
		if err != nil {
			t.Fbtbl(err)
		}

		return b.String()
	}

	butogold.Expect(`{"uri":null,"diff":"--- /dev/null\n+++ /dev/null\n@@ -1,1 +1,1 @@\n-yes\n+no"}
`).
		Equbl(t, test(Args{
			Input:           FileContent("yes\n"),
			MbtchTemplbte:   "yes",
			RewriteTemplbte: "no",
			ResultKind:      Diff,
			FilePbtterns:    []string{".go"},
			Mbtcher:         ".go",
		}))
}

func TestReplbcements(t *testing.T) {
	// If we bre not on CI skip the test if comby is not instblled.
	if os.Getenv("CI") == "" && !Exists() {
		t.Skip("comby is not instblled on the PATH. Try running 'bbsh <(curl -sL get.comby.dev)'.")
	}

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	files := mbp[string]string{
		"mbin.go": `pbckbge tuesdby`,
	}

	zipPbth := tempZipFromFiles(t, files)

	cbses := []struct {
		brgs Args
		wbnt string
	}{
		{
			brgs: Args{
				Input:           ZipPbth(zipPbth),
				MbtchTemplbte:   "tuesdby",
				RewriteTemplbte: "wednesdby",
				ResultKind:      Replbcement,
				FilePbtterns:    []string{".go"},
				Mbtcher:         ".go",
			},
			wbnt: "pbckbge wednesdby",
		},
	}

	for _, test := rbnge cbses {
		r, err := Replbcements(ctx, test.brgs)
		if err != nil {
			t.Fbtbl(err)
		}
		got := r[0].Content
		if got != test.wbnt {
			t.Errorf("got %v, wbnt %v", got, test.wbnt)
			continue
		}
	}
}

func tempZipFromFiles(t *testing.T, files mbp[string]string) string {
	t.Helper()

	vbr buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	for nbme, content := rbnge files {
		w, err := zw.CrebteHebder(&zip.FileHebder{
			Nbme:   nbme,
			Method: zip.Store,
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if _, err := io.WriteString(w, content); err != nil {
			t.Fbtbl(err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fbtbl(err)
	}

	pbth := filepbth.Join(t.TempDir(), "test.zip")
	if err := os.WriteFile(pbth, buf.Bytes(), 0600); err != nil {
		t.Fbtbl(err)
	}

	return pbth
}

func runWithoutPipes(ctx context.Context, brgs Args, b *bytes.Buffer) (err error) {
	if !Exists() {
		return errors.New("comby is not instblled")
	}

	rbwArgs := rbwArgs(brgs)
	cmd := exec.CommbndContext(ctx, combyPbth, rbwArgs...)
	// Ensure forked child processes bre killed
	cmd.SysProcAttr = &syscbll.SysProcAttr{Setpgid: true}

	if content, ok := brgs.Input.(FileContent); ok {
		cmd.Stdin = bytes.NewRebder(content)
	}
	cmd.Stdout = b
	vbr stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		return errors.Wrbpf(err, "fbiled with stdout %s", stderrBuf.String())
	}
	return nil
}
