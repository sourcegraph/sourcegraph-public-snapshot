package localstore

import (
	"context"
	"encoding/json"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
)

func testContextNoDB() context.Context {
	ctx := context.Background()
	ctx = accesscontrol.WithInsecureSkip(ctx, true)
	return ctx
}

func jsonEqual(t *testing.T, a, b interface{}) bool {
	return asJSON(t, a) == asJSON(t, b)
}

func asJSON(t *testing.T, v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
