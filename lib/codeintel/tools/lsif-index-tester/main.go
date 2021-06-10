package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/conversion"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/lsif/validation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/semantic"
)

type projectResult struct {
	name       string
	success    bool
	usage      usageStats
	output     string
	testResult testSuiteResult
}

type usageStats struct {
	// Memory usage in kilobytes by child process.
	memory int64
}

type passedTest struct {
	Name string
}

type failedTest struct {
	Name string
	Diff string
}

type testFileResult struct {
	Name   string
	Passed []passedTest
	Failed []failedTest
}

type testSuiteResult struct {
	FileResults []testFileResult
}

var directory string
var raw_indexer string
var debug bool

// TODO: Do more monitoring of the process.
// var monitor bool

func logFatal(msg string, args ...interface{}) {
	log15.Error(msg, args...)
	os.Exit(1)
}

func main() {
	flag.StringVar(&directory, "dir", ".", "The directory to run the test harness over")
	flag.StringVar(&raw_indexer, "indexer", "", "The name of the indexer that you want to test")
	flag.BoolVar(&debug, "debug", false, "Enable debugging")

	flag.Parse()

	if debug {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlDebug, log15.StdoutHandler))
	} else {
		log15.Root().SetHandler(log15.LvlFilterHandler(log15.LvlError, log15.StdoutHandler))
	}

	if raw_indexer == "" {
		logFatal("Indexer is required. Pass with --indexer")
	}

	var indexer = strings.Split(raw_indexer, " ")

	log15.Info("Starting Execution: ", "directory", directory, "indexer", raw_indexer)

	if err := testDirectory(context.Background(), indexer, directory); err != nil {
		logFatal("Failed with", "err", err)
	}
	log15.Info("Tests passed:", "directory", directory, "indexer", raw_indexer)
}

func testDirectory(ctx context.Context, indexer []string, directory string) error {
	files, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	type channelResult struct {
		result projectResult
		err    error
	}

	resultChan := make(chan channelResult, len(files))
	var wg sync.WaitGroup

	for _, f := range files {
		wg.Add(1)

		go func(name string) {
			defer wg.Done()

			projResult, err := testProject(ctx, indexer, path.Join(directory, name), name)

			resultChan <- channelResult{
				result: projResult,
				err:    err,
			}
		}(f.Name())

	}

	wg.Wait()
	close(resultChan)

	allPassed := true
	for chanResult := range resultChan {
		if chanResult.err != nil {
			log15.Warn("Failed to run test. Got err:", "error", chanResult.err)
			continue
		}

		for _, fileResult := range chanResult.result.testResult.FileResults {
			if len(fileResult.Failed) > 0 {
				allPassed = false
				for _, failed := range fileResult.Failed {
					fmt.Printf("Failed test: File: %s, Name: %s\nDiff: %s\n", fileResult.Name, failed.Name, failed.Diff)
				}
			}
		}
	}

	if !allPassed {
		return errors.New("Failed some tests. Try again later :)")
	}

	return nil
}

func testProject(ctx context.Context, indexer []string, project, name string) (projectResult, error) {
	output, err := setupProject(project)
	if err != nil {
		return projectResult{name: name, success: false, output: string(output)}, err
	}

	log15.Debug("... Completed setup project", "command", indexer)

	result, err := runIndexer(ctx, indexer, project, name)
	if err != nil {
		return projectResult{
			name:    name,
			success: false,
			output:  result.output,
		}, err
	}

	log15.Debug("... \t Resource Usage:", "usage", result.usage)

	if err := validateDump(project); err != nil {
		return projectResult{}, err
	}
	log15.Debug("... Validated dump.lsif")

	bundle, err := readBundle(project)
	if err != nil {
		return projectResult{}, err
	}

	testResult, err := validateTestCases(project, bundle)
	if err != nil {
		return projectResult{}, err
	}

	return projectResult{
		name:       name,
		success:    true,
		usage:      result.usage,
		output:     string(output),
		testResult: testResult,
	}, nil
}

func setupProject(directory string) ([]byte, error) {
	cmd := exec.Command("./setup_indexer.sh")
	cmd.Dir = directory

	return cmd.CombinedOutput()
}

func runIndexer(ctx context.Context, indexer []string, directory, name string) (projectResult, error) {
	command := indexer[0]
	args := indexer[1:]

	log15.Debug("... Generating dump.lsif")
	cmd := exec.Command(command, args...)
	cmd.Dir = directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		return projectResult{}, err
	}

	sysUsage := cmd.ProcessState.SysUsage()
	mem, _ := MaxMemoryInKB(sysUsage)

	return projectResult{
		name:    name,
		success: false,
		usage:   usageStats{memory: mem},
		output:  string(output),
	}, err
}

func validateDump(directory string) error {
	dumpFile, err := os.Open(filepath.Join(directory, "dump.lsif"))
	if err != nil {
		return err
	}

	ctx := validation.NewValidationContext()
	validator := &validation.Validator{Context: ctx}

	if err := validator.Validate(dumpFile); err != nil {
		return err
	}

	if len(ctx.Errors) > 0 {
		msg := fmt.Sprintf("Detected %d errors", len(ctx.Errors))
		for i, err := range ctx.Errors {
			msg += fmt.Sprintf("%d. %s", i, err)
		}
		return errors.New(msg)
	}

	return nil
}

func validateTestCases(directory string, bundle *semantic.GroupedBundleDataMaps) (testSuiteResult, error) {
	testFiles, err := os.ReadDir(filepath.Join(directory, "lsif_tests"))
	if err != nil {
		if os.IsNotExist(err) {
			log15.Warn("No lsif test directory exists here", "directory", directory)
			return testSuiteResult{}, nil
		}

		return testSuiteResult{}, err
	}

	fileResults := make([]testFileResult, len(testFiles))
	for _, file := range testFiles {
		if testFileExtension := filepath.Ext(file.Name()); testFileExtension != "json" {
			continue
		}

		fileResult, err := runOneTestFile(filepath.Join(directory, "lsif_tests", file.Name()), bundle)
		if err != nil {
			logFatal("Had an error while we do the test file", "err", err)
		}

		fileResults = append(fileResults, fileResult)
	}

	return testSuiteResult{FileResults: fileResults}, nil
}

func runOneTestFile(file string, bundle *semantic.GroupedBundleDataMaps) (testFileResult, error) {
	doc, err := os.ReadFile(file)
	if err != nil {
		return testFileResult{}, errors.Wrap(err, "Failed to read file")
	}

	var testCase LsifTest
	if err := json.Unmarshal(doc, &testCase); err != nil {
		return testFileResult{}, errors.Wrap(err, "Malformed JSON")
	}

	fileResult := testFileResult{Name: file}

	for _, definitionRequest := range testCase.Definitions {
		path := definitionRequest.Request.TextDocument
		line := definitionRequest.Request.Position.Line
		character := definitionRequest.Request.Position.Character

		results, err := semantic.Query(bundle, path, line, character)
		if err != nil {
			return testFileResult{}, err
		}

		// TODO: We probably can have more than one result and have that make sense...
		//		should allow testing that
		if len(results) > 1 {
			return testFileResult{}, errors.New("Had too many results")
		} else if len(results) == 0 {
			return testFileResult{}, errors.New("Found no results")
		}

		definitions := results[0].Definitions

		if len(definitions) > 1 {
			logFatal("Had too many definitions", "definitions", definitions)
		} else if len(definitions) == 0 {
			logFatal("Found no definitions", "definitions", definitions)
		}

		response := transformLocationToResponse(definitions[0])
		if diff := cmp.Diff(response, definitionRequest.Response); diff != "" {
			fileResult.Failed = append(fileResult.Failed, failedTest{
				Name: definitionRequest.Name,
				Diff: diff,
			})
		} else {
			fileResult.Passed = append(fileResult.Passed, passedTest{
				Name: definitionRequest.Name,
			})
		}
	}

	return fileResult, nil
}

func transformLocationToResponse(location semantic.LocationData) DefinitionResponse {
	return DefinitionResponse{
		TextDocument: location.URI,
		Range: Range{
			Start: Position{
				Line:      location.StartLine,
				Character: location.StartCharacter,
			},
			End: Position{
				Line:      location.EndLine,
				Character: location.EndCharacter,
			},
		},
	}

}
func readBundle(root string) (*semantic.GroupedBundleDataMaps, error) {
	bundle, err := conversion.CorrelateLocalGit(context.Background(), path.Join(root, "dump.lsif"), root)
	if err != nil {
		return nil, err
	}

	return semantic.GroupedBundleDataChansToMaps(bundle), nil
}
