package main

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Test struct {
	PreExistingCacheEntries []execution.AfterStepResult
	BatchSpecInput          string
	ExpectedCacheEntries    []execution.AfterStepResult
	ExpectedChangesetSpecs  []types.ChangesetSpec
	ExpectedState           BatchSpec
}

type BatchSpec struct {
	State          string
	FailureMessage string
}

func ReadTest(testName string) error {
	test := Test{}
	dir := filepath.Join("testdata", testName)
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("test is not a directory")
	}
	file, err := os.Open(filepath.Join(dir, "cache_data.json"))
	if err != nil {
		return err
	}

	if err := json.NewDecoder(file).Decode(&test.PreExistingCacheEntries); err != nil {
		file.Close()
		return err
	}
	file.Close()

	// TODO: Read additional data.

	return nil
}

func RunTest(db database.DB, test Test) error {

}
