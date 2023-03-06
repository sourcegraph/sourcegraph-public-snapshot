package graphql

import (
	"encoding/base64"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/shared"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
)

func TestMakeGetUploadsOptions(t *testing.T) {
	opts, err := makeGetUploadsOptions(&resolverstubs.LSIFRepositoryUploadsQueryArgs{
		LSIFUploadsQueryArgs: &resolverstubs.LSIFUploadsQueryArgs{
			ConnectionArgs: resolverstubs.ConnectionArgs{
				First: intPtr(5),
			},
			Query:           strPtr("q"),
			State:           strPtr("s"),
			IsLatestForRepo: boolPtr(true),
			After:           resolverstubs.EncodeIntCursor(intPtr(25)).EndCursor(),
		},
		RepositoryID: graphql.ID(base64.StdEncoding.EncodeToString([]byte("Repo:50"))),
	})
	if err != nil {
		t.Fatalf("unexpected error making options: %s", err)
	}

	expected := shared.GetUploadsOptions{
		RepositoryID: 50,
		State:        "s",
		Term:         "q",
		VisibleAtTip: true,
		Limit:        5,
		Offset:       25,
		AllowExpired: true,
	}
	if diff := cmp.Diff(expected, opts); diff != "" {
		t.Errorf("unexpected opts (-want +got):\n%s", diff)
	}
}

func TestMakeGetUploadsOptionsDefaults(t *testing.T) {
	opts, err := makeGetUploadsOptions(&resolverstubs.LSIFRepositoryUploadsQueryArgs{
		LSIFUploadsQueryArgs: &resolverstubs.LSIFUploadsQueryArgs{},
	})
	if err != nil {
		t.Fatalf("unexpected error making options: %s", err)
	}

	expected := shared.GetUploadsOptions{
		RepositoryID: 0,
		State:        "",
		Term:         "",
		VisibleAtTip: false,
		Limit:        DefaultUploadPageSize,
		Offset:       0,
		AllowExpired: true,
	}
	if diff := cmp.Diff(expected, opts); diff != "" {
		t.Errorf("unexpected opts (-want +got):\n%s", diff)
	}
}

func TestCursor(t *testing.T) {
	expected := "test"
	pageInfo := EncodeCursor(&expected)

	if !pageInfo.HasNextPage() {
		t.Fatalf("expected next page")
	}
	if pageInfo.EndCursor() == nil {
		t.Fatalf("unexpected nil cursor")
	}

	value, err := DecodeCursor(pageInfo.EndCursor())
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != expected {
		t.Errorf("unexpected decoded cursor. want=%s have=%s", expected, value)
	}
}

func TestCursorEmpty(t *testing.T) {
	pageInfo := EncodeCursor(nil)

	if pageInfo.HasNextPage() {
		t.Errorf("unexpected next page")
	}
	if pageInfo.EndCursor() != nil {
		t.Errorf("unexpected encoded cursor: %s", *pageInfo.EndCursor())
	}

	value, err := DecodeCursor(nil)
	if err != nil {
		t.Fatalf("unexpected error decoding cursor: %s", err)
	}
	if value != "" {
		t.Errorf("unexpected decoded cursor: %s", value)
	}
}

func TestUploadID(t *testing.T) {
	expected := int64(42)
	value, err := unmarshalLSIFUploadGQLID(marshalLSIFUploadGQLID(expected))
	if err != nil {
		t.Fatalf("unexpected error marshalling id: %s", err)
	}
	if value != expected {
		t.Errorf("unexpected id. have=%d want=%d", expected, value)
	}
}

func TestUnmarshalUploadIDString(t *testing.T) {
	expected := int64(42)
	id := graphql.ID(base64.StdEncoding.EncodeToString([]byte(`LSIFUpload:"42"`)))
	value, err := unmarshalLSIFUploadGQLID(id)
	if err != nil {
		t.Fatalf("unexpected error marshalling id: %s", err)
	}
	if value != expected {
		t.Errorf("unexpected id. have=%d want=%d", expected, value)
	}
}

func TestDerefInt32(t *testing.T) {
	expected := 42
	expected32 := int32(expected)

	if val := derefInt32(nil, expected); val != expected {
		t.Errorf("unexpected value. want=%d have=%d", expected, val)
	}
	if val := derefInt32(&expected32, expected); val != expected {
		t.Errorf("unexpected value. want=%d have=%d", expected, val)
	}
}

func TestDerefString(t *testing.T) {
	expected := "foo"

	if val := derefString(nil, expected); val != expected {
		t.Errorf("unexpected value. want=%s have=%s", expected, val)
	}
	if val := derefString(&expected, ""); val != expected {
		t.Errorf("unexpected value. want=%s have=%s", expected, val)
	}
	if val := derefString(&expected, expected); val != expected {
		t.Errorf("unexpected value. want=%s have=%s", expected, val)
	}
}

func TestDerefBool(t *testing.T) {
	if val := derefBool(nil, true); !val {
		t.Errorf("unexpected value. want=%v have=%v", true, val)
	}
	if val := derefBool(nil, false); val {
		t.Errorf("unexpected value. want=%v have=%v", false, val)
	}

	pVal := true
	if val := derefBool(&pVal, true); !val {
		t.Errorf("unexpected value. want=%v have=%v", true, val)
	}
	if val := derefBool(&pVal, false); !val {
		t.Errorf("unexpected value. want=%v have=%v", false, val)
	}

	pVal = false
	if val := derefBool(&pVal, true); val {
		t.Errorf("unexpected value. want=%v have=%v", true, val)
	}
	if val := derefBool(&pVal, false); val {
		t.Errorf("unexpected value. want=%v have=%v", false, val)
	}
}
