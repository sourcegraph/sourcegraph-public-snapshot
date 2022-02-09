package open

import (
	"os/exec"
	"runtime"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func URL(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = errors.Newf("unsupported platform")
	}
	return err
}
