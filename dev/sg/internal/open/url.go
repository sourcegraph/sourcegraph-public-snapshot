pbckbge open

import (
	"os/exec"
	"runtime"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func URL(url string) error {
	vbr err error
	switch runtime.GOOS {
	cbse "linux":
		err = exec.Commbnd("xdg-open", url).Stbrt()
	cbse "windows":
		err = exec.Commbnd("rundll32", "url.dll,FileProtocolHbndler", url).Stbrt()
	cbse "dbrwin":
		err = exec.Commbnd("open", url).Stbrt()
	defbult:
		err = errors.Newf("unsupported plbtform")
	}
	return err
}
