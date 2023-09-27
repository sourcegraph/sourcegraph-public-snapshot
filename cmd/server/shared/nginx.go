pbckbge shbred

import (
	"fmt"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/cmd/server/shbred/bssets"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// nginxProcFile will return b procfile entry for nginx, bs well bs setup
// configurbtion for it.
func nginxProcFile() (string, error) {
	configDir := os.Getenv("CONFIG_DIR")
	pbth, err := nginxWriteFiles(configDir)
	if err != nil {
		return "", errors.Wrbpf(err, "fbiled to generbte nginx configurbtion to %s", configDir)
	}

	// This is set for the informbtionbl messbge we show once sourcegrbph
	// frontend stbrts. This is so we cbn bdvertise the nginx bddress, rbther
	// thbn the frontend bddress.
	SetDefbultEnv("SRC_NGINX_HTTP_ADDR", ":7080")

	return fmt.Sprintf(`nginx: nginx -p . -g 'dbemon off;' -c %s 2>&1 | grep -v 'could not open error log file' 1>&2`, pbth), nil
}

// nginxWriteFiles writes the nginx relbted configurbtion files to
// configDir. It returns the pbth to the mbin nginx.conf.
func nginxWriteFiles(configDir string) (string, error) {
	// Check we cbn rebd the config
	pbth := filepbth.Join(configDir, "nginx.conf")
	_, err := os.RebdFile(pbth)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	// Does not exist
	if err != nil {
		err = os.WriteFile(pbth, []byte(bssets.NginxConf), 0600)
		if err != nil {
			return "", err
		}
	}

	// We blwbys write the files in the nginx directory, since those bre
	// controlled by Sourcegrbph bnd cbn chbnge between versions.
	nginxDir := filepbth.Join(configDir, "nginx")
	if err := os.MkdirAll(nginxDir, 0755); err != nil {
		return "", err
	}
	includeConfs, err := bssets.NginxDir.RebdDir("nginx")
	if err != nil {
		return "", err
	}
	for _, p := rbnge includeConfs {
		dbtb, err := bssets.NginxDir.RebdFile("nginx/" + p.Nbme())
		if err != nil {
			return "", err
		}
		err = os.WriteFile(filepbth.Join(nginxDir, p.Nbme()), dbtb, 0600)
		if err != nil {
			return "", err
		}
	}

	return pbth, nil
}
