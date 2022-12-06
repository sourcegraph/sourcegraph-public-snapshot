package adr

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/grafana/regexp"
)

type ArchitectureDecisionRecord struct {
	Number int
	Title  string
	Date   time.Time

	// The following are set if ADR is read or created
	Path     string
	BasePath string
}

// DocsiteURL returns a link to this ADR in docs.sourcegraph.com
func (r ArchitectureDecisionRecord) DocsiteURL() string {
	cleanedName := r.BasePath[:len(r.BasePath)-len(filepath.Ext(r.BasePath))]
	return fmt.Sprintf("https://docs.sourcegraph.com/dev/adr/" + cleanedName)
}

// List parses all ADRs and returns them in read order.
func List(adrDir string) ([]ArchitectureDecisionRecord, error) {
	var adrs []ArchitectureDecisionRecord
	return adrs, VisitAll(adrDir, func(adr ArchitectureDecisionRecord) error {
		adrs = append(adrs, adr)
		return nil
	})
}

var (
	// Matches for ADRs with filename format ${timestamp}-${name}.md
	adrFilenameRegexp = regexp.MustCompile(`^(\d+)-.+\.md`)
	// Matches for Markdown headers
	markdownHeaderRegexp = regexp.MustCompile(`#\s+(\d+)\.\s+(.*)$`)
)

// VisitAll applies visit on all ADRs.
//
// Must be kept in sync with the generator in Create.
func VisitAll(adrDir string, visit func(adr ArchitectureDecisionRecord) error) error {
	// nolint:staticcheck,gosimple
	return filepath.WalkDir(adrDir, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		// Ensure this file matches the ADR format
		filenameMatch := adrFilenameRegexp.FindAllStringSubmatch(entry.Name(), 1)
		if filenameMatch == nil {
			return nil
		}

		// Parse the timestamp - we can ignore the err because we know from the regexp
		// it's only digits.
		ts, _ := strconv.Atoi(filenameMatch[0][1])
		date := time.Unix(int64(ts), 0)

		// Look for more details in the file contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		s := bufio.NewScanner(file)
		for s.Scan() {
			headerMatches := markdownHeaderRegexp.FindAllStringSubmatch(s.Text(), 1)
			// We only care about the first header match, so process it to get ADR details
			// and exit.
			if len(headerMatches) > 0 {
				// We can ignore the err because we know from the regexp it's only digits.
				number, _ := strconv.Atoi(headerMatches[0][1])
				// Title is after the number
				title := headerMatches[0][2]

				// Pass to visit
				if err := visit(ArchitectureDecisionRecord{
					Title:  title,
					Number: number,
					Date:   date,

					Path:     path,
					BasePath: entry.Name(),
				}); err != nil {
					return err
				}
				break
			}
		}

		return nil
	})
}

// Create generates an ADR template file.
//
// Must be kept in sync with the parser in VisitAll.
func Create(adrDir string, adr *ArchitectureDecisionRecord) error {
	fileName := fmt.Sprintf("%d-%s.md", adr.Date.Unix(), sanitizeADRName(adr.Title))
	f, err := os.Create(filepath.Join(adrDir, fileName))
	if err != nil {
		return err
	}

	// Update the ADR
	adr.Path = f.Name()
	adr.BasePath = filepath.Base(adr.Path)

	// Write header
	fmt.Fprintf(f, "# %d. %s\n\n", adr.Number, adr.Title)
	fmt.Fprintf(f, "Date: %s\n\n", adr.Date.Format("2006-01-02"))

	// Create sections
	fmt.Fprint(f, "## Context\n\nTODO\n\n")
	fmt.Fprint(f, "## Decision\n\nTODO\n\n")
	fmt.Fprint(f, "## Consequences\n\nTODO\n")

	// Save file
	return f.Sync()
}
