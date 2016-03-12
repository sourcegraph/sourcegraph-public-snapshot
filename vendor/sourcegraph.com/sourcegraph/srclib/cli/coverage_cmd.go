package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"strings"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/cvg"
	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/grapher"
	"sourcegraph.com/sourcegraph/srclib/plan"
)

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

var codeExts = []string{".go", ".java", ".py", ".rb", ".cpp", ".ts", ".cs"} // codeExt lists all file extensions that indicate a code file we want to cover
var codeExts_ = make(map[string]struct{})

func init() {
	for _, ext := range codeExts {
		codeExts_[ext] = struct{}{}
	}
}

func coverage(repo *Repo) (*cvg.Coverage, error) {
	lineSep := []byte{'\n'}
	codeFileData := make(map[string]*codeFileDatum)
	log.Printf(repo.RootDir)
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
		if _, isCodeFile := codeExts_[filepath.Ext(path)]; isCodeFile {
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			loc := bytes.Count(b, lineSep)
			codeFileData[path] = &codeFileDatum{LoC: loc}
		}
		return nil
	})

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

	for _, rule_ := range mf.Rules {
		rule, ok := rule_.(*grapher.GraphUnitRule)
		if !ok {
			continue
		}

		var data graph.Output
		if err := readJSONFileFS(bdfs, rule.Target(), &data); err != nil {
			if err == errEmptyJSONFile {
				log.Printf("Warning: the JSON file is empty for unit %s %s.", rule.Unit.Type, rule.Unit.Name)
				continue
			}
			if os.IsNotExist(err) {
				log.Printf("Warning: no build data for unit %s %s.", rule.Unit.Type, rule.Unit.Name)
				continue
			}
			return nil, fmt.Errorf("error reading JSON file %s for unit %s %s: %s", rule.Target(), rule.Unit.Type, rule.Unit.Name, err)
		}

		defKeys := make(map[graph.DefKey]struct{})
		for _, def := range data.Defs {
			defKeys[def.DefKey] = struct{}{}
		}
		var validRefs []*graph.Ref
		for _, ref := range data.Refs {
			if datum, exists := codeFileData[ref.File]; exists {
				datum.NumRefs++

				if ref.DefRepo != "" {
					validRefs = append(validRefs, ref)
					datum.NumRefsValid++
				} else if _, defExists := defKeys[ref.DefKey()]; defExists {
					validRefs = append(validRefs, ref)
					datum.NumRefsValid++
				}
			}
		}

		for _, def := range data.Defs {
			if datum, exists := codeFileData[def.File]; exists {
				datum.NumDefs++
			}
		}
	}

	var fileTokThresh float32 = 0.7
	numIndexedFiles := 0
	numDefs, numRefs, numRefsValid := 0, 0, 0
	loc := 0
	var uncoveredFiles []string
	for file, datum := range codeFileData {
		loc += datum.LoC
		numDefs += datum.NumDefs
		numRefs += datum.NumRefs
		numRefsValid += datum.NumRefsValid
		if float32(datum.NumDefs+datum.NumRefsValid)/float32(datum.LoC) > fileTokThresh {
			numIndexedFiles++
		} else {
			uncoveredFiles = append(uncoveredFiles, file)
		}
	}

	return &cvg.Coverage{
		FileScore:      float32(numIndexedFiles) / float32(len(codeFileData)),
		RefScore:       float32(numRefsValid) / float32(numRefs),
		TokDensity:     float32(numDefs+numRefs) / float32(loc),
		UncoveredFiles: uncoveredFiles,
	}, nil
}
