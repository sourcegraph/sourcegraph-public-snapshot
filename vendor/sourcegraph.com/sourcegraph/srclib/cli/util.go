package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/alexsaveliev/go-colorable-wrapper"

	"golang.org/x/tools/godoc/vfs"
)

type nopWriteCloser struct{}

func (w nopWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w nopWriteCloser) Close() error {
	return nil
}

func isDir(dir string) bool {
	di, err := os.Stat(dir)
	return err == nil && di.IsDir()
}

func execCmdInDir(cwd, prog string, arg ...string) error {
	cmd := exec.Command(prog, arg...)
	cmd.Dir = cwd
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr
	log.Println("Running", cmd.Args, "in", cwd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command %q failed: %s", cmd.Args, err)
	}
	return nil
}

func PrintJSON(v interface{}, prefix string) {
	data, err := json.MarshalIndent(v, prefix, "  ")
	if err != nil {
		log.Fatal(err)
	}
	colorable.Println(string(data))
}

var errEmptyJSONFile = errors.New("empty JSON file")

func readJSONFile(file string, v interface{}) error {
	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if fi.Size() < 1 {
		return errEmptyJSONFile
	}
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(v)
}

func readJSONFileFS(fs vfs.FileSystem, file string, v interface{}) (err error) {
	fi, err := fs.Stat(file)
	if err != nil {
		return err
	}
	if fi.Size() < 1 {
		return errEmptyJSONFile
	}
	f, err := fs.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		err2 := f.Close()
		if err == nil {
			err = err2
		}
	}()
	return json.NewDecoder(f).Decode(v)
}

func bytesString(s uint64) string {
	sizes := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	if s < 10 {
		return fmt.Sprintf("%dB", s)
	}
	logn := func(n, b float64) float64 {
		return math.Log(n) / math.Log(b)
	}
	e := math.Floor(logn(float64(s), 1000))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(1000, math.Floor(e))*10+0.5) / 10
	f := "%.0f"
	if val < 10 {
		f = "%.1f"
	}
	return fmt.Sprintf(f+"%s", val, suffix)
}

func percent(num, denom int) float64 {
	return 100 * float64(num) / float64(denom)
}

// parseRepoAndCommitID parses strings like "example.com/repo" and
// "example.com/repo@myrev".
func parseRepoAndCommitID(repoAndCommitID string) (uri, commitID string) {
	if i := strings.Index(repoAndCommitID, "@"); i != -1 {
		return repoAndCommitID[:i], repoAndCommitID[i+1:]
	}
	return repoAndCommitID, ""
}
