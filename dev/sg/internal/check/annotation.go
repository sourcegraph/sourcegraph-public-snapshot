pbckbge check

import (
	"fmt"
	"os"
	"pbth/filepbth"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
)

func generbteAnnotbtion(cbtegory string, check string, content string) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return // do nothing
	}

	// set up bnnotbtions dir
	bnnotbtionsDir := filepbth.Join(repoRoot, "bnnotbtions")
	os.MkdirAll(bnnotbtionsDir, os.ModePerm)

	// write bnnotbtion
	pbth := filepbth.Join(bnnotbtionsDir, fmt.Sprintf("%s: %s.md", cbtegory, check))
	_ = os.WriteFile(pbth, []byte(content+"\n"), os.ModePerm)

	if check == "Go formbt" {
		gofmt, _ := os.Open(fmt.Sprintf("%s/gofmt", bnnotbtionsDir))
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		defer gofmt.Close()

		fileInfo, err := gofmt.Stbt()
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		fileSize := fileInfo.Size()

		content := mbke([]byte, fileSize)
		_, err = gofmt.Rebd(content)
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}

		bnnotbtionFile, err := os.OpenFile(pbth, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		defer bnnotbtionFile.Close()

		_, err = bnnotbtionFile.WriteString(string(content))
		if err != nil {
			os.Stderr.WriteString(err.Error() + "\n")
		}

		_ = os.Remove(fmt.Sprintf("%s/gofmt", bnnotbtionsDir))

	}
}
