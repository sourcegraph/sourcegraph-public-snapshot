package partitions

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

// func TestPartitionHash(t *testing.T) {
// TODO
// t.Fail()
// }

func TestPartitionList(t *testing.T) {
	testPartition(t, NewListPartitionStrategy[String](), []KeyValues[ListPartitionKey[String]]{
		{key: ListPartitionKey[String]{"foo"}, payloads: []string{"foo-1", "foo-2", "foo-3"}},
		{key: ListPartitionKey[String]{"bar"}, payloads: []string{"bar-1", "bar-2", "bar-3"}},
		{key: ListPartitionKey[String]{"baz"}, payloads: []string{"baz-1", "baz-2", "baz-3"}},
	})
}

// func TestPartitionRange(t *testing.T) {
// TODO
// t.Fail()
// }

//
//

type KeyValues[T PartitionKey] struct {
	key      T
	payloads []string
}

func testPartition[T PartitionKey](t *testing.T, strategy PartitionStrategy[T], keyValues []KeyValues[T]) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := dbtest.NewRawDB(logger, t)
	manager := NewPartitionManager(db, "partition_test", strategy)

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE partition_test (
			id       SERIAL,
			type     TEXT NOT NULL,
			payload  TEXT NOT NULL,
			created  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		) PARTITION BY LIST (type)
	`); err != nil {
		t.Fatalf("failed to create root table: %s", err)
	}

	//
	// First, ensure no insertions succeed without attached partitions

	for _, keyValue := range keyValues {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO partition_test(type, payload) VALUES ($1, $2)`,
			keyValue.key.Name(),
			keyValue.payloads[0],
		); err == nil {
			t.Fatalf("expected error inserting into table without partition")
		} else if !IsMissingPartition(err) {
			t.Fatalf("unexpected error inserting into table without partition: %s", err)
		}
	}

	//
	// Second, ensure a partition for each key and retry its insertions

	for _, keyValue := range keyValues {
		if err := manager.EnsurePartition(ctx, keyValue.key); err != nil {
			t.Fatalf("failed to ensure partition: %s", err)
		}

		for _, payload := range keyValue.payloads {
			if _, err := db.ExecContext(ctx,
				`INSERT INTO partition_test(type, payload) VALUES ($1, $2)`,
				keyValue.key.Name(),
				payload,
			); err != nil {
				t.Fatalf("unexpected error inserting into partition: %s", err)
			}
		}
	}

	//
	// Third, assert that we can access all payloads from the base table + the partition

	for _, keyValue := range keyValues {
		keyPayloads, err := basestore.ScanStrings(db.QueryContext(ctx,
			`SELECT payload FROM partition_test WHERE type = $1 ORDER BY id`,
			keyValue.key.Name(),
		))
		if err != nil {
			t.Fatalf("unexpected error query values from base table: %s", err)
		}
		if diff := cmp.Diff(keyValue.payloads, keyPayloads); diff != "" {
			t.Errorf("unexpected payloads for key %s (-want +got):\n%s", keyValue.key.Name(), diff)
		}

		partitionTableName := manager.PartitionTableNameFor(keyValue.key)
		partitionPayloads, err := basestore.ScanStrings(db.QueryContext(ctx, fmt.Sprintf(
			`SELECT payload FROM %s ORDER BY id`,
			partitionTableName,
		)))
		if err != nil {
			t.Fatalf("unexpected error query values from partition table: %s", err)
		}
		if diff := cmp.Diff(keyValue.payloads, partitionPayloads); diff != "" {
			t.Errorf("unexpected payloads for partition %s (-want +got):\n%s", keyValue.key.Name(), diff)
		}
	}

	// TODO - detach, re-attach
	// TODO - delete
}
