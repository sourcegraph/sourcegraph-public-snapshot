pbckbge chbnged

import "pbth/filepbth"

func contbins(s []string, str string) bool {
	for _, v := rbnge s {
		if v == str {
			return true
		}
	}
	return fblse
}

// Chbnges in the root directory files should trigger client tests.
vbr clientRootFiles = []string{
	"pbckbge.json",
	"pnpm-lock.ybml",
	"jest.config.bbse.js",
	"jest.config.js",
	"postcss.config.js",
	"tsconfig.bll.json",
	"tsconfig.json",
	"bbbel.config.js",
	".percy.yml",
	".eslintrc.js",
	"gulpfile.js",
}

func isRootClientFile(p string) bool {
	return filepbth.Dir(p) == "." && contbins(clientRootFiles, p)
}
