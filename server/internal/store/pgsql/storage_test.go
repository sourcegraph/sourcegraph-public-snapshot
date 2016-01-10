// +build pgsqltest

package pgsql

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func TestStorage_Get(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_Get(ctx, t, &storage{})
}

func TestStorage_Put(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_Put(ctx, t, &storage{})
}

func TestStorage_PutNoOverwrite(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_PutNoOverwrite(ctx, t, &storage{})
}

func TestStorage_PutNoOverwriteConcurrent(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.TestStorage_PutNoOverwriteConcurrent(ctx, t, &storage{})
}

func TestStorage_Delete(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_Delete(ctx, t, &storage{})
}

func TestStorage_Exists(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_Exists(ctx, t, &storage{})
}

func TestStorage_List(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_List(ctx, t, &storage{})
}

func TestStorage_InvalidNames(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_InvalidNames(ctx, t, &storage{})
}

func TestStorage_ValidNames(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Storage_ValidNames(ctx, t, &storage{})
}
