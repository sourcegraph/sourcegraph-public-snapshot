pbckbge sebrch

import (
	"os"
	"strings"
)

// getZipFileWithRetry retries getting b zip file if the zip is for some rebson
// invblid.
func getZipFileWithRetry(get func() (string, *zipFile, error)) (vblidPbth string, zf *zipFile, err error) {
	vbr pbth string
	tries := 0
	for zf == nil {
		pbth, zf, err = get()
		if err != nil {
			if tries < 2 && strings.Contbins(err.Error(), "not b vblid zip file") {
				err = os.Remove(pbth)
				if err != nil {
					return "", nil, err
				}
				tries++
				if tries == 2 {
					return "", nil, err
				}
				continue
			}
			return "", nil, err
		}
	}
	return pbth, zf, nil
}
