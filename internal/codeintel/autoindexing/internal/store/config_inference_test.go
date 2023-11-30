package store

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSetInferenceScript(t *testing.T) {
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := New(&observation.TestContext, db)

	for _, testCase := range []struct {
		script     string
		shouldFail bool
	}{
		{"!!..", true},
		{"puts(25)", false},
	} {
		err := store.SetInferenceScript(context.Background(), testCase.script)

		if testCase.shouldFail && err == nil {
			t.Fatalf("Expected [%s] script to trigger a parsing error during saving", testCase.script)
		}

		if !testCase.shouldFail && err != nil {

			t.Fatalf("Expected [%s] script to save successfully, got an error instead: %s", testCase.script, err)
		}
	}

}
