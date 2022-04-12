package main

import (
	"fmt"
	"testing"
)

func TestCheckForCommit(t *testing.T) {
	tt := []struct {
		Version   string
		CommitLog []Commit
		Result    bool
	}{
		{
			Version:   "1234567",
			CommitLog: []Commit{{Sha: "7025d04"}, {Sha: "1bbdfb1"}, {Sha: "bfe1b89"}},
			Result:    false,
		},
		{
			Version:   "7025d04",
			CommitLog: []Commit{{Sha: "7025d04"}, {Sha: "1bbdfb1"}, {Sha: "bfe1b89"}},
			Result:    true,
		},
	}

	for i, test := range tt {
		testName := fmt.Sprintf("%v", i)
		t.Run(testName, func(t *testing.T) {
			result := checkForCommit(test.Version, test.CommitLog)
			if result != test.Result {
				t.Errorf("Invalid result. Expected: %v Got %v", test.Result, result)
			}
		})
	}
}
