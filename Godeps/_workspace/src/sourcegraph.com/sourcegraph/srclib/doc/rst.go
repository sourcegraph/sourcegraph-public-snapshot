package doc

import (
	"bytes"
	"os"
	"os/exec"
)

var rst2html string

func init() {
	rst2html = os.Getenv("RST2HTML")
	if rst2html == "" {
		rst2html, _ = exec.LookPath("rst2html.py")
	}
	if rst2html == "" {
		rst2html, _ = exec.LookPath("rst2html")
	}
}

func ReStructuredTextToHTML(rst []byte) ([]byte, error) {
	cmd := exec.Command(rst2html, "--quiet")
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	in.Write(rst)
	in.Close()
	html, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	start := bytes.Index(html, []byte("<body>")) + len("<body>")
	end := bytes.Index(html, []byte("</body>"))
	return bytes.TrimSpace(html[start:end]), nil
}
