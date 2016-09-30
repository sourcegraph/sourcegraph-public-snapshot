package localstore

import (
	"testing"

	"gopkg.in/gorp.v1"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil"
)

func TestIsPQErrorUniqueViolation(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, done := testContext()
	defer done()
	dbutil.Transact(appDBH(ctx), func(tx gorp.SqlExecutor) error {
		tx.Exec(`CREATE TEMPORARY TABLE pq_test_tmp (id int PRIMARY KEY) ON COMMIT DROP`)
		if isPQErrorUniqueViolation(nil) {
			t.Errorf("nil errors should be false")
		}
		_, err := tx.Exec(`INSERT INTO pq_test_tmp(id) VALUES(1)`)
		if err != nil {
			t.Fatalf("Initial insert should not fail: %s", err)
		}
		_, err = tx.Exec(`INSERT INTO pq_test_tmp(id) VALUES(1)`)
		if err == nil {
			t.Fatalf("Second insert should fail")
		}
		if !isPQErrorUniqueViolation(err) {
			t.Fatalf("Expected error to be a unique_violation: %s", err)
		}
		return err
	})
}
