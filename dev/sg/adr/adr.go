package adr

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
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

var (
	adrFileRegexp        = regexp.MustCompile(`^(\d+)-.+\.md`)
	markdownHeaderRegexp = regexp.MustCompile(`#\s+(\d+)\.\s+(.*)$`)
)

func List(adrDir string) ([]ArchitectureDecisionRecord, error) {
	var adrs []ArchitectureDecisionRecord
	return adrs, VisitAll(adrDir, func(adr ArchitectureDecisionRecord) error {
		adrs = append(adrs, adr)
		return nil
	})
}

// VisitAll applies visit on all ADRs.
//
// Must be kept in sync with the generator in Create.
func VisitAll(adrDir string, visit func(adr ArchitectureDecisionRecord) error) error {
	return filepath.WalkDir(adrDir, func(path string, entry fs.DirEntry, err error) error {
		if entry.IsDir() {
			return nil
		}

		if !adrFileRegexp.MatchString(entry.Name()) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		m := adrFileRegexp.FindAllStringSubmatch(entry.Name(), 1)
		ts, _ := strconv.Atoi(m[0][1]) // We can ignore the err because we know from the regexp it's only digits.

		s := bufio.NewScanner(bytes.NewReader(b))
		var title string
		var number int
		for s.Scan() {
			matches := markdownHeaderRegexp.FindAllStringSubmatch(s.Text(), 1)
			if len(matches) > 0 {
				number, _ = strconv.Atoi(matches[0][1]) // We can ignore the err because we know from the regexp it's only digits.
				title = matches[0][2]
				if err := visit(ArchitectureDecisionRecord{
					Title:  title,
					Number: number,
					Date:   time.Unix(int64(ts), 0),

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

var nonAlphaNumericOrDash = regexp.MustCompile("[^a-z0-9-]+")

func sanitizeADRName(name string) string {
	return nonAlphaNumericOrDash.ReplaceAllString(
		strings.ReplaceAll(strings.ToLower(name), " ", "-"), "",
	)
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
