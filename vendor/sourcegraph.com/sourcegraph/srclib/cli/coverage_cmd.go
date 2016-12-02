package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"

	"strings"
	"unicode"

	"text/scanner"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/cvg"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/grapher"
	"sourcegraph.com/sourcegraph/srclib/plan"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

const fileTokThresh float64 = 0.7

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		_, err := cli.AddCommand("coverage",
			"srclib coverage",
			"compute approximate amount of code successfully analyzed by srclib",
			&coverageCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
	})
}

type codeFileDatum struct {
	LoC          int
	NumRefs      int
	NumDefs      int
	NumRefsValid int
	Language     string
	Seen         bool
}

type CoverageCmd struct {
}

var coverageCmd CoverageCmd

func (c *CoverageCmd) Execute(args []string) error {
	repo, err := OpenLocalRepo()
	if err != nil {
		return err
	}

	cvg, err := coverage(repo)
	if err != nil {
		return err
	}

	out, err := json.MarshalIndent(cvg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))

	return nil
}

var langToExts = map[string][]string{
	"Go":          {".go"},
	"Java":        {".java"},
	"Python":      {".py"},
	"Ruby":        {".rb"},
	"C++":         {".cpp", ".cc", ".cxx", ".c++"},
	"TypeScript":  {".ts"},
	"C#":          {".cs"},
	"JavaScript":  {".js"},
	"PHP":         {".php"},
	"Objective-C": {".m", ".mm"},
}
var extToLang map[string]string

func init() {
	extToLang = make(map[string]string)
	for lang, exts := range langToExts {
		for _, ext := range exts {
			extToLang[ext] = lang
		}
	}
}

func coverage(repo *Repo) (map[string]*cvg.Coverage, error) {
	// Gather file data
	codeFileData := make(map[string]*codeFileDatum) // data for each file needed to compute coverage
	filepath.Walk(repo.RootDir, func(path string, info os.FileInfo, err error) error {
		if filepath.IsAbs(path) {
			var err error
			path, err = filepath.Rel(repo.RootDir, path)
			if err != nil {
				return err
			}
		}

		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir // don't search hidden directories
			}
			return nil
		}

		path = filepath.ToSlash(path)

		ext := strings.ToLower(filepath.Ext(path))
		if lang, isCodeFile := extToLang[ext]; isCodeFile {

			// omitting special files (auto-generated, temporary, ...)
			if shouldIgnoreFile(path, lang) {
				return nil
			}

			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			loc := numLines(b)
			codeFileData[path] = &codeFileDatum{LoC: loc, Language: lang}
		}
		return nil
	})

	// Gather ref/def data for each file
	bdfs, err := GetBuildDataFS(repo.CommitID)
	if err != nil {
		return nil, err
	}
	treeConfig, err := config.ReadCached(bdfs)
	if err != nil {
		return nil, fmt.Errorf("error calling config.ReadCached: %s", err)
	}
	mf, err := plan.CreateMakefile(".", nil, "", treeConfig)
	if err != nil {
		return nil, fmt.Errorf("error calling plan.Makefile: %s", err)
	}

	defKeys := make(map[graph.DefKey]struct{})
	data := make([]graph.Output, 0, len(mf.Rules))

	parseGraphData := func(graphFile string, sourceUnit *unit.SourceUnit) error {
		var item graph.Output
		if err := readJSONFileFS(bdfs, graphFile, &item); err != nil {
			if err == errEmptyJSONFile {
				log.Printf("Warning: the JSON file is empty for unit %s %s.", sourceUnit.Type, sourceUnit.Name)
				return nil
			}
			if os.IsNotExist(err) {
				log.Printf("Warning: no build data for unit %s %s.", sourceUnit.Type, sourceUnit.Name)
				return nil
			}
			return fmt.Errorf("error reading JSON file %s for unit %s %s: %s", graphFile, sourceUnit.Type, sourceUnit.Name, err)
		}
		data = append(data, item)

		for _, file := range sourceUnit.Files {
			if datum, exists := codeFileData[file]; exists {
				datum.Seen = true
			}
		}

		for _, def := range item.Defs {
			defKeys[def.DefKey] = struct{}{}
		}

		return nil
	}

	for _, rule_ := range mf.Rules {
		switch rule := rule_.(type) {
		case *grapher.GraphUnitRule:
			if err := parseGraphData(rule.Target(), rule.Unit); err != nil {
				return nil, err
			}
		case *grapher.GraphMultiUnitsRule:
			for target, sourceUnit := range rule.Targets() {
				if err := parseGraphData(target, sourceUnit); err != nil {
					return nil, err
				}
			}
		}
	}

	missingKeys := make(map[graph.DefKey]struct{})

	for _, item := range data {
		var validRefs []*graph.Ref
		for _, ref := range item.Refs {
			if datum, exists := codeFileData[ref.File]; exists {
				datum.NumRefs++

				if ref.DefUnitType == "URL" || ref.DefRepo != "" {
					validRefs = append(validRefs, ref)
					datum.NumRefsValid++
				} else if _, defExists := defKeys[ref.DefKey()]; defExists {
					validRefs = append(validRefs, ref)
					datum.NumRefsValid++
				} else if GlobalOpt.Verbose {
					if _, reported := missingKeys[ref.DefKey()]; !reported {
						missingKeys[ref.DefKey()] = struct{}{}
						sample := ref.DefKey().Path
						candidates := make([]graph.DefKey, 0, 1)
						for key := range defKeys {
							if key.Path == sample {
								candidates = append(candidates, key)
							}
						}
						log.Printf("No matching def for %s, candidates are %v", ref.String(), candidates)
					}
				}
			}
		}

		for _, def := range item.Defs {
			if datum, exists := codeFileData[def.File]; exists {
				datum.NumDefs++
			}
		}
	}

	// Compute coverage from per-file data
	type langStats struct {
		numFiles          int
		numIndexedFiles   int
		numDefs           int
		numRefs           int
		numRefsValid      int
		uncoveredFiles    []string
		undiscoveredFiles []string
		loc               int
	}
	stats := make(map[string]*langStats)
	for file, datum := range codeFileData {
		if _, exist := stats[datum.Language]; !exist {
			stats[datum.Language] = &langStats{}
		}

		s := stats[datum.Language]
		s.loc += datum.LoC
		s.numDefs += datum.NumDefs
		s.numRefs += datum.NumRefs
		s.numRefsValid += datum.NumRefsValid
		if datum.Seen {
			// this file is listed in the source unit and found by the scanner
			s.numFiles++
			density := float64(datum.NumDefs+datum.NumRefsValid) / float64(datum.LoC)
			if density > fileTokThresh {
				s.numIndexedFiles++
			} else {
				if GlobalOpt.Verbose {
					log.Printf("Uncovered file %s - density: %f, defs: %d, refs: %d, lines of code: %d",
						file, density, datum.NumDefs, datum.NumRefsValid, datum.LoC)
				}
				s.uncoveredFiles = append(s.uncoveredFiles, file)
			}
		} else {
			// this file is not listed in the source unit but found by the scanner
			if GlobalOpt.Verbose {
				log.Printf("Undiscovered file %s", file)
			}
			s.undiscoveredFiles = append(s.undiscoveredFiles, file)
		}
	}

	cov := make(map[string]*cvg.Coverage)
	for lang, s := range stats {
		cov[lang] = &cvg.Coverage{
			FileScore:         divideSentinel(float64(s.numIndexedFiles), float64(s.numFiles), -1),
			RefScore:          divideSentinel(float64(s.numRefsValid), float64(s.numRefs), -1),
			TokDensity:        divideSentinel(float64(s.numDefs+s.numRefs), float64(s.loc), -1),
			UncoveredFiles:    s.uncoveredFiles,
			UndiscoveredFiles: s.undiscoveredFiles,
		}
	}
	return cov, nil
}

// shouldIgnoreFile returns true if file denoted by the given path should be
// ignored when scanning for files
func shouldIgnoreFile(filename, language string) bool {
	basename := filepath.Base(filename)
	switch {
	case language == "Go":
		return basename == "doc.go"
	case language == "Java":
		// ignoring Andoid auto-generated stuff
		return basename == "R.java" || basename == "BuildConfig.java"
	case language == "JavaScript":
		// ignoring everything in the node_modules directory
		return strings.HasPrefix(filename, "node_modules/")
	}
	return false
}

func divideSentinel(x, y, sentinel float64) float64 {
	q := x / y
	if math.IsNaN(q) {
		return sentinel
	}
	return q
}

// numLines counts the number of lines that
// - are not blank
// - do not look like comments
func numLines(data []byte) int {

	data = stripCode(data)

	len := len(data)
	if len == 0 {
		return 0
	}

	count := 0
	start := 0

	pos := bytes.IndexByte(data[start:], '\n')
	for pos != -1 && start < len {
		l := data[start : start+pos+1]
		if isNotBlank(l) {
			count++
		}
		start += pos + 1
		pos = bytes.IndexByte(data[start:], '\n')
	}
	if start < len && isNotBlank(data[start:]) {
		count++
	}

	return count
}

// stripCode strips all but idents from the data
func stripCode(data []byte) []byte {

	ret := make([]byte, 0, len(data))
	s := &scanner.Scanner{}
	s.Init(bytes.NewReader(data))
	s.Error = func(_ *scanner.Scanner, _ string) {}
	s.Mode = s.Mode ^ scanner.SkipComments

	offset := 0

	tok := s.Scan()
	for tok != scanner.EOF {
		pos := s.Pos()
		// there may be cases when Go scanner is inaccurate for some languages
		// but it's ok because the purpose of this method is to strip eerything
		// that doesn't look like ident so it probably won't be recognized by
		// toolchain as a ref or def span. Of course there may be exceptions
		// such as C++ operators, Groovy's GStrings, Python comments and so on,
		// but we think that removing Go-style idents from source code gives
		// pretty accurate resutls when counting number of lines that may contain
		// ref/def spans
		if tok != scanner.Ident {
			ret = append(ret, data[offset:pos.Offset-len(s.TokenText())]...)
		} else {
			ret = append(ret, data[offset:pos.Offset]...)
		}
		offset = pos.Offset
		tok = s.Scan()
	}

	return append(ret, data[offset:]...)
}

// isNotBlank returns true if data contains at least one not-whitespace character
func isNotBlank(data []byte) bool {
	for _, r := range data {
		if !unicode.IsSpace(rune(r)) {
			return true
		}
	}
	return false
}
