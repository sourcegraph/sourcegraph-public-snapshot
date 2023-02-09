package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetDependencies(t *testing.T) {
	store := populateTestStore(t)

	dependencies, err := store.GetDependencies(context.Background(), []int{testSCIPUploadID})
	if err != nil {
		t.Fatalf("unexpected error %s", err)
	}

	expectedDependencies := []DependencyDescription{
		{Manager: "npm", Name: "@sinonjs/fake-timers", Version: "7.1.2"},
		{Manager: "npm", Name: "@sourcegraph/extension-api-stubs", Version: "1.6.1"},
		{Manager: "npm", Name: "@types/fs-extra", Version: "9.0.13"},
		{Manager: "npm", Name: "@types/lodash", Version: "4.14.178"},
		{Manager: "npm", Name: "@types/lru-cache", Version: "5.1.1"},
		{Manager: "npm", Name: "@types/mocha", Version: "9.0.0"},
		{Manager: "npm", Name: "@types/mock-require", Version: "2.0.1"},
		{Manager: "npm", Name: "@types/mz", Version: "2.7.4"},
		{Manager: "npm", Name: "@types/node", Version: "14.17.15"},
		{Manager: "npm", Name: "@types/sinon", Version: "10.0.4"},
		{Manager: "npm", Name: "@types/yargs", Version: "17.0.2"},
		{Manager: "npm", Name: "code-intel-extensions", Version: "0.0.0-DEVELOPMENT"},
		{Manager: "npm", Name: "fast-json-stable-stringify", Version: "2.1.0"},
		{Manager: "npm", Name: "js-base64", Version: "3.7.1"},
		{Manager: "npm", Name: "rxjs", Version: "6.6.7"},
		{Manager: "npm", Name: "sourcegraph", Version: "25.5.0"},
		{Manager: "npm", Name: "tagged-template-noop", Version: "2.1.01"},
		{Manager: "npm", Name: "template", Version: "0.0.0-DEVELOPMENT"},
		{Manager: "npm", Name: "typescript", Version: "4.9.3"},
	}
	if diff := cmp.Diff(dependencies, expectedDependencies); diff != "" {
		t.Errorf("unexpected dependencies (-want +got):\n%s", diff)
	}
}
