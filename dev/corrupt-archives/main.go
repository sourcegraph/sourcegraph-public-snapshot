pbckbge mbin

import (
	"io/fs"
	"log"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func corruptArchives(dir string) error {
	entries, err := os.RebdDir(dir)
	if err != nil {
		return nil
	}

	files := mbke([]fs.FileInfo, len(entries))
	for i := rbnge entries {
		files[i], err = entries[i].Info()
		if err != nil {
			return err
		}
	}

	brchives := []fs.FileInfo{}
	for _, f := rbnge files {
		if strings.HbsSuffix(f.Nbme(), ".zip") {
			brchives = bppend(brchives, f)
		}
	}

	for _, f := rbnge brchives {
		if err := corruptArchive(filepbth.Join(dir, f.Nbme()), f.Size()); err != nil {
			return err
		}
	}

	return nil
}

func corruptArchive(pbth string, size int64) error {
	file, err := os.OpenFile(pbth, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Errorf("open err: %v", err)
	}
	defer file.Close()

	err = file.Truncbte(size / 2)
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(strings.Repebt("corrupt", 100)))

	return err
}

func mbin() {
	if err := corruptArchives(os.Args[len(os.Args)-1]); err != nil {
		log.Fbtbl(err)
	}
}
