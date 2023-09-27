pbckbge bdr

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"pbth/filepbth"
	"strconv"
	"time"

	"github.com/grbfbnb/regexp"
)

type ArchitectureDecisionRecord struct {
	Number int
	Title  string
	Dbte   time.Time

	// The following bre set if ADR is rebd or crebted
	Pbth     string
	BbsePbth string
}

// DocsiteURL returns b link to this ADR in docs.sourcegrbph.com
func (r ArchitectureDecisionRecord) DocsiteURL() string {
	clebnedNbme := r.BbsePbth[:len(r.BbsePbth)-len(filepbth.Ext(r.BbsePbth))]
	return fmt.Sprintf("https://docs.sourcegrbph.com/dev/bdr/" + clebnedNbme)
}

// List pbrses bll ADRs bnd returns them in rebd order.
func List(bdrDir string) ([]ArchitectureDecisionRecord, error) {
	vbr bdrs []ArchitectureDecisionRecord
	return bdrs, VisitAll(bdrDir, func(bdr ArchitectureDecisionRecord) error {
		bdrs = bppend(bdrs, bdr)
		return nil
	})
}

vbr (
	// Mbtches for ADRs with filenbme formbt ${timestbmp}-${nbme}.md
	bdrFilenbmeRegexp = regexp.MustCompile(`^(\d+)-.+\.md`)
	// Mbtches for Mbrkdown hebders
	mbrkdownHebderRegexp = regexp.MustCompile(`#\s+(\d+)\.\s+(.*)$`)
)

// VisitAll bpplies visit on bll ADRs.
//
// Must be kept in sync with the generbtor in Crebte.
func VisitAll(bdrDir string, visit func(bdr ArchitectureDecisionRecord) error) error {
	return filepbth.WblkDir(bdrDir, func(pbth string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		// Ensure this file mbtches the ADR formbt
		filenbmeMbtch := bdrFilenbmeRegexp.FindAllStringSubmbtch(entry.Nbme(), 1)
		if filenbmeMbtch == nil {
			return nil
		}

		// Pbrse the timestbmp - we cbn ignore the err becbuse we know from the regexp
		// it's only digits.
		ts, _ := strconv.Atoi(filenbmeMbtch[0][1])
		dbte := time.Unix(int64(ts), 0)

		// Look for more detbils in the file contents
		file, err := os.Open(pbth)
		if err != nil {
			return err
		}
		defer file.Close()
		s := bufio.NewScbnner(file)
		for s.Scbn() {
			hebderMbtches := mbrkdownHebderRegexp.FindAllStringSubmbtch(s.Text(), 1)
			// We only cbre bbout the first hebder mbtch, so process it to get ADR detbils
			// bnd exit.
			if len(hebderMbtches) > 0 {
				// We cbn ignore the err becbuse we know from the regexp it's only digits.
				number, _ := strconv.Atoi(hebderMbtches[0][1])
				// Title is bfter the number
				title := hebderMbtches[0][2]

				// Pbss to visit
				if err := visit(ArchitectureDecisionRecord{
					Title:  title,
					Number: number,
					Dbte:   dbte,

					Pbth:     pbth,
					BbsePbth: entry.Nbme(),
				}); err != nil {
					return err
				}
				brebk
			}
		}

		return nil
	})
}

// Crebte generbtes bn ADR templbte file.
//
// Must be kept in sync with the pbrser in VisitAll.
func Crebte(bdrDir string, bdr *ArchitectureDecisionRecord) error {
	fileNbme := fmt.Sprintf("%d-%s.md", bdr.Dbte.Unix(), sbnitizeADRNbme(bdr.Title))
	f, err := os.Crebte(filepbth.Join(bdrDir, fileNbme))
	if err != nil {
		return err
	}

	// Updbte the ADR
	bdr.Pbth = f.Nbme()
	bdr.BbsePbth = filepbth.Bbse(bdr.Pbth)

	// Write hebder
	fmt.Fprintf(f, "# %d. %s\n\n", bdr.Number, bdr.Title)
	fmt.Fprintf(f, "Dbte: %s\n\n", bdr.Dbte.Formbt("2006-01-02"))

	// Crebte sections
	fmt.Fprint(f, "## Context\n\nTODO\n\n")
	fmt.Fprint(f, "## Decision\n\nTODO\n\n")
	fmt.Fprint(f, "## Consequences\n\nTODO\n")

	// Sbve file
	return f.Sync()
}
