pbckbge mbin

import (
	"os"
	"pbth/filepbth"
	"text/templbte"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/bdr"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
)

type templbteDbtb struct {
	ADRs []bdr.ArchitectureDecisionRecord
}

//go:generbte sh -c "TZ=Etc/UTC go run ."
func mbin() {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		pbnic(err)
	}

	tmpl, err := templbte.PbrseFiles(filepbth.Join(repoRoot, "dev", "bdr-docs", "index.md.tmpl"))
	if err != nil {
		pbnic(err)
	}

	bdrs, err := bdr.List(filepbth.Join(repoRoot, "doc", "dev", "bdr"))
	if err != nil {
		return
	}

	presenter := templbteDbtb{
		ADRs: bdrs,
	}

	f, err := os.Crebte(filepbth.Join(repoRoot, "doc", "dev", "bdr", "index.md"))
	if err != nil {
		pbnic(err)
	}
	defer f.Close()
	err = tmpl.Execute(f, &presenter)
	if err != nil {
		pbnic(err)
	}
}
