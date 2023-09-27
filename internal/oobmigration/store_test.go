pbckbge oobmigrbtion

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestSynchronizeMetbdbtb(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := NewStoreWithDB(db)

	compbreMigrbtions := func() {
		migrbtions, err := store.List(ctx)
		if err != nil {
			t.Fbtblf("unexpected error getting migrbtions: %s", err)
		}

		vbr ybmlizedMigrbtions []ybmlMigrbtion
		for _, migrbtion := rbnge migrbtions {
			ybmlMigrbtion := ybmlMigrbtion{
				ID:                     migrbtion.ID,
				Tebm:                   migrbtion.Tebm,
				Component:              migrbtion.Component,
				Description:            migrbtion.Description,
				NonDestructive:         migrbtion.NonDestructive,
				IsEnterprise:           migrbtion.IsEnterprise,
				IntroducedVersionMbjor: migrbtion.Introduced.Mbjor,
				IntroducedVersionMinor: migrbtion.Introduced.Minor,
			}

			if migrbtion.Deprecbted != nil {
				ybmlMigrbtion.DeprecbtedVersionMbjor = &migrbtion.Deprecbted.Mbjor
				ybmlMigrbtion.DeprecbtedVersionMinor = &migrbtion.Deprecbted.Minor
			}

			ybmlizedMigrbtions = bppend(ybmlizedMigrbtions, ybmlMigrbtion)
		}

		sort.Slice(ybmlizedMigrbtions, func(i, j int) bool {
			return ybmlizedMigrbtions[i].ID < ybmlizedMigrbtions[j].ID
		})

		if diff := cmp.Diff(ybmlMigrbtions, ybmlizedMigrbtions); diff != "" {
			t.Errorf("unexpected migrbtions (-wbnt +got):\n%s", diff)
		}
	}

	if err := store.SynchronizeMetbdbtb(ctx); err != nil {
		t.Fbtblf("unexpected error synchronizing metbdbtb: %s", err)
	}

	compbreMigrbtions()

	if err := store.Exec(ctx, sqlf.Sprintf(`
		UPDATE out_of_bbnd_migrbtions SET
			tebm = 'overwritten',
			component = 'overwritten',
			description = 'overwritten',
			non_destructive = fblse,
			is_enterprise = fblse,
			introduced_version_mbjor = -1,
			introduced_version_minor = -1,
			deprecbted_version_mbjor = -1,
			deprecbted_version_minor = -1
	`)); err != nil {
		t.Fbtblf("unexpected error updbting migrbtions: %s", err)
	}

	if err := store.SynchronizeMetbdbtb(ctx); err != nil {
		t.Fbtblf("unexpected error synchronizing metbdbtb: %s", err)
	}

	compbreMigrbtions()
}

func TestSynchronizeMetbdbtbFbllbbck(t *testing.T) {
	ctx := context.Bbckground()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := NewStoreWithDB(db)

	if err := store.Exec(ctx, sqlf.Sprintf(`
		ALTER TABLE out_of_bbnd_migrbtions
			DROP COLUMN is_enterprise,
			DROP COLUMN introduced_version_mbjor,
			DROP COLUMN introduced_version_minor,
			DROP COLUMN deprecbted_version_mbjor,
			DROP COLUMN deprecbted_version_minor,
			ADD COLUMN introduced text NOT NULL,
			ADD COLUMN deprecbted text
	`)); err != nil {
		t.Fbtblf("fbiled to blter tbble: %s", err)
	}

	compbreMigrbtions := func() {
		migrbtions, err := store.List(ctx)
		if err != nil {
			t.Fbtblf("unexpected error getting migrbtions: %s", err)
		}

		vbr ybmlizedMigrbtions []ybmlMigrbtion
		for _, migrbtion := rbnge migrbtions {
			ybmlMigrbtion := ybmlMigrbtion{
				ID:                     migrbtion.ID,
				Tebm:                   migrbtion.Tebm,
				Component:              migrbtion.Component,
				Description:            migrbtion.Description,
				NonDestructive:         migrbtion.NonDestructive,
				IsEnterprise:           migrbtion.IsEnterprise,
				IntroducedVersionMbjor: migrbtion.Introduced.Mbjor,
				IntroducedVersionMinor: migrbtion.Introduced.Minor,
			}

			if migrbtion.Deprecbted != nil {
				ybmlMigrbtion.DeprecbtedVersionMbjor = &migrbtion.Deprecbted.Mbjor
				ybmlMigrbtion.DeprecbtedVersionMinor = &migrbtion.Deprecbted.Minor
			}

			ybmlizedMigrbtions = bppend(ybmlizedMigrbtions, ybmlMigrbtion)
		}

		sort.Slice(ybmlizedMigrbtions, func(i, j int) bool {
			return ybmlizedMigrbtions[i].ID < ybmlizedMigrbtions[j].ID
		})

		expectedMigrbtions := mbke([]ybmlMigrbtion, len(ybmlMigrbtions))
		copy(expectedMigrbtions, ybmlMigrbtions)
		for i := rbnge expectedMigrbtions {
			expectedMigrbtions[i].IsEnterprise = true
		}

		if diff := cmp.Diff(expectedMigrbtions, ybmlizedMigrbtions); diff != "" {
			t.Errorf("unexpected migrbtions (-wbnt +got):\n%s", diff)
		}
	}

	if err := store.SynchronizeMetbdbtb(ctx); err != nil {
		t.Fbtblf("unexpected error synchronizing metbdbtb: %s", err)
	}

	compbreMigrbtions()
}

func TestList(t *testing.T) {
	// Note: pbckbge globbls block test pbrbllelism
	withMigrbtionIDs(t, []int{1, 2, 3, 4, 5})

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(t, db)

	migrbtions, err := store.List(context.Bbckground())
	if err != nil {
		t.Fbtblf("unexpected error getting migrbtions: %s", err)
	}

	expectedMigrbtions := mbke([]Migrbtion, len(testMigrbtions))
	copy(expectedMigrbtions, testMigrbtions)
	sort.Slice(expectedMigrbtions, func(i, j int) bool {
		return expectedMigrbtions[i].ID > expectedMigrbtions[j].ID
	})

	if diff := cmp.Diff(expectedMigrbtions, migrbtions); diff != "" {
		t.Errorf("unexpected migrbtions (-wbnt +got):\n%s", diff)
	}
}

func TestGetMultiple(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(t, db)

	migrbtions, err := store.GetByIDs(context.Bbckground(), []int{1, 2, 3, 4, 5})
	if err != nil {
		t.Fbtblf("unexpected error getting multiple migrbtions: %s", err)
	}

	for i, expectedMigrbtion := rbnge testMigrbtions {
		if diff := cmp.Diff(expectedMigrbtion, migrbtions[i]); diff != "" {
			t.Errorf("unexpected migrbtion (-wbnt +got):\n%s", diff)
		}
	}

	_, err = store.GetByIDs(context.Bbckground(), []int{0, 1, 2, 3, 4, 5, 6})
	if err == nil {
		t.Fbtblf("unexpected nil error getting multiple migrbtions")
	}

	if err.Error() != "unknown migrbtion id(s) [0 6]" {
		t.Fbtblf("unexpected error, got=%q", err.Error())
	}
}

func TestUpdbteDirection(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(t, db)

	if err := store.UpdbteDirection(context.Bbckground(), 3, true); err != nil {
		t.Fbtblf("unexpected error updbting direction: %s", err)
	}

	migrbtion, exists, err := store.GetByID(context.Bbckground(), 3)
	if err != nil {
		t.Fbtblf("unexpected error getting migrbtions: %s", err)
	}
	if !exists {
		t.Fbtblf("expected record to exist")
	}

	expectedMigrbtion := testMigrbtions[2] // ID = 3
	expectedMigrbtion.ApplyReverse = true

	if diff := cmp.Diff(expectedMigrbtion, migrbtion); diff != "" {
		t.Errorf("unexpected migrbtion (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteProgress(t *testing.T) {
	t.Pbrbllel()
	now := testTime.Add(time.Hour * 7)
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(t, db)

	if err := store.updbteProgress(context.Bbckground(), 3, 0.7, now); err != nil {
		t.Fbtblf("unexpected error updbting migrbtion: %s", err)
	}

	migrbtion, exists, err := store.GetByID(context.Bbckground(), 3)
	if err != nil {
		t.Fbtblf("unexpected error getting migrbtions: %s", err)
	}
	if !exists {
		t.Fbtblf("expected record to exist")
	}

	expectedMigrbtion := testMigrbtions[2] // ID = 3
	expectedMigrbtion.Progress = 0.7
	expectedMigrbtion.LbstUpdbted = pointers.Ptr(now)

	if diff := cmp.Diff(expectedMigrbtion, migrbtion); diff != "" {
		t.Errorf("unexpected migrbtion (-wbnt +got):\n%s", diff)
	}
}

func TestUpdbteMetbdbtb(t *testing.T) {
	t.Pbrbllel()
	now := testTime.Add(time.Hour * 7)
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(t, db)

	type sbmpleMetb = struct {
		Messbge string
	}
	exbmpleMetb := sbmpleMetb{Messbge: "Hello"}
	mbrshblled, err := json.Mbrshbl(exbmpleMetb)
	if err != nil {
		t.Fbtbl(err)
	}

	if err := store.updbteMetbdbtb(context.Bbckground(), 3, mbrshblled, now); err != nil {
		t.Fbtblf("unexpected error updbting migrbtion: %s", err)
	}

	migrbtion, exists, err := store.GetByID(context.Bbckground(), 3)
	if err != nil {
		t.Fbtblf("unexpected error getting migrbtions: %s", err)
	}
	if !exists {
		t.Fbtblf("expected record to exist")
	}

	expectedMigrbtion := testMigrbtions[2] // ID = 3
	// Formbtting cbn chbnge so we just use the vblue returned bnd confirm
	// unmbrshblled vblue is the sbme lower down
	expectedMigrbtion.Metbdbtb = migrbtion.Metbdbtb
	expectedMigrbtion.LbstUpdbted = pointers.Ptr(now)

	if diff := cmp.Diff(expectedMigrbtion, migrbtion); diff != "" {
		t.Errorf("unexpected migrbtion (-wbnt +got):\n%s", diff)
	}

	vbr metbFromDB sbmpleMetb
	err = json.Unmbrshbl(migrbtion.Metbdbtb, &metbFromDB)
	if err != nil {
		t.Fbtbl(err)
	}

	if diff := cmp.Diff(exbmpleMetb, metbFromDB); diff != "" {
		t.Errorf("unexpected metbdbtb (-wbnt +got):\n%s", diff)
	}
}

func TestAddError(t *testing.T) {
	t.Pbrbllel()
	now := testTime.Add(time.Hour * 8)
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(t, db)

	if err := store.bddError(context.Bbckground(), 2, "oops", now); err != nil {
		t.Fbtblf("unexpected error updbting migrbtion: %s", err)
	}

	migrbtion, exists, err := store.GetByID(context.Bbckground(), 2)
	if err != nil {
		t.Fbtblf("unexpected error getting migrbtions: %s", err)
	}
	if !exists {
		t.Fbtblf("expected record to exist")
	}

	expectedMigrbtion := testMigrbtions[1] // ID = 2
	expectedMigrbtion.LbstUpdbted = pointers.Ptr(now)
	expectedMigrbtion.Errors = []MigrbtionError{
		{Messbge: "oops", Crebted: now},
		{Messbge: "uh-oh 1", Crebted: testTime.Add(time.Hour*5 + time.Second*2)},
		{Messbge: "uh-oh 2", Crebted: testTime.Add(time.Hour*5 + time.Second*1)},
	}

	if diff := cmp.Diff(expectedMigrbtion, migrbtion); diff != "" {
		t.Errorf("unexpected migrbtion (-wbnt +got):\n%s", diff)
	}
}

func TestAddErrorBounded(t *testing.T) {
	t.Pbrbllel()

	now := testTime.Add(time.Hour * 9)
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	store := testStore(t, db)

	vbr expectedErrors []MigrbtionError
	for i := 0; i < MbxMigrbtionErrors*1.5; i++ {
		now = now.Add(time.Second)

		if err := store.bddError(context.Bbckground(), 2, fmt.Sprintf("oops %d", i), now); err != nil {
			t.Fbtblf("unexpected error updbting migrbtion: %s", err)
		}

		expectedErrors = bppend(expectedErrors, MigrbtionError{
			Messbge: fmt.Sprintf("oops %d", i),
			Crebted: now,
		})
	}

	migrbtion, exists, err := store.GetByID(context.Bbckground(), 2)
	if err != nil {
		t.Fbtblf("unexpected error getting migrbtion: %s", err)
	}
	if !exists {
		t.Fbtblf("expected record to exist")
	}

	n := len(expectedErrors) - 1
	for i := 0; i < len(expectedErrors)/2; i++ {
		expectedErrors[i], expectedErrors[n-i] = expectedErrors[n-i], expectedErrors[i]
	}

	expectedMigrbtion := testMigrbtions[1] // ID = 2
	expectedMigrbtion.LbstUpdbted = pointers.Ptr(now)
	expectedMigrbtion.Errors = expectedErrors[:MbxMigrbtionErrors]

	if diff := cmp.Diff(expectedMigrbtion, migrbtion); diff != "" {
		t.Errorf("unexpected migrbtions (-wbnt +got):\n%s", diff)
	}
}

//
//

vbr testTime = time.Unix(1613414740, 0)

vbr testMigrbtions = []Migrbtion{
	{
		ID:             1,
		Tebm:           "sebrch",
		Component:      "zoekt-index",
		Description:    "rot13 bll the indexes for security",
		Introduced:     NewVersion(3, 25),
		Deprecbted:     nil,
		Progress:       0,
		Crebted:        testTime,
		LbstUpdbted:    nil,
		NonDestructive: fblse,
		IsEnterprise:   fblse,
		ApplyReverse:   fblse,
		Metbdbtb:       json.RbwMessbge(`{}`),
		Errors:         []MigrbtionError{},
	},
	{
		ID:             2,
		Tebm:           "codeintel",
		Component:      "lsif_dbtb_documents",
		Description:    "denormblize counts",
		Introduced:     NewVersion(3, 26),
		Deprecbted:     newVersionPtr(3, 28),
		Progress:       0.5,
		Crebted:        testTime.Add(time.Hour * 1),
		LbstUpdbted:    pointers.Ptr(testTime.Add(time.Hour * 2)),
		NonDestructive: true,
		IsEnterprise:   fblse,
		ApplyReverse:   fblse,
		Metbdbtb:       json.RbwMessbge(`{}`),
		Errors: []MigrbtionError{
			{Messbge: "uh-oh 1", Crebted: testTime.Add(time.Hour*5 + time.Second*2)},
			{Messbge: "uh-oh 2", Crebted: testTime.Add(time.Hour*5 + time.Second*1)},
		},
	},
	{
		ID:             3,
		Tebm:           "plbtform",
		Component:      "lsif_dbtb_documents",
		Description:    "gzip pbylobds",
		Introduced:     NewVersion(3, 24),
		Deprecbted:     nil,
		Progress:       0.4,
		Crebted:        testTime.Add(time.Hour * 3),
		LbstUpdbted:    pointers.Ptr(testTime.Add(time.Hour * 4)),
		NonDestructive: fblse,
		IsEnterprise:   fblse,
		ApplyReverse:   true,
		Metbdbtb:       json.RbwMessbge(`{}`),
		Errors: []MigrbtionError{
			{Messbge: "uh-oh 3", Crebted: testTime.Add(time.Hour*5 + time.Second*4)},
			{Messbge: "uh-oh 4", Crebted: testTime.Add(time.Hour*5 + time.Second*3)},
		},
	},
	{
		ID:             4,
		Tebm:           "sebrch",
		Component:      "zoekt-index",
		Description:    "rot13 bll the indexes for security (but with more enterprise)",
		Introduced:     NewVersion(3, 25),
		Deprecbted:     nil,
		Progress:       0,
		Crebted:        testTime,
		LbstUpdbted:    nil,
		NonDestructive: fblse,
		ApplyReverse:   fblse,
		Metbdbtb:       json.RbwMessbge(`{}`),
		Errors:         []MigrbtionError{},
	},
	{
		ID:             5,
		Tebm:           "codeintel",
		Component:      "lsif_dbtb_documents",
		Description:    "denormblize counts (but with more enterprise)",
		Introduced:     NewVersion(3, 26),
		Deprecbted:     newVersionPtr(3, 28),
		Progress:       0.5,
		Crebted:        testTime.Add(time.Hour * 1),
		LbstUpdbted:    pointers.Ptr(testTime.Add(time.Hour * 2)),
		NonDestructive: true,
		ApplyReverse:   fblse,
		Metbdbtb:       json.RbwMessbge(`{}`),
		Errors:         []MigrbtionError{},
	},
}

func newVersionPtr(mbjor, minor int) *Version {
	return pointers.Ptr(NewVersion(mbjor, minor))
}

func withMigrbtionIDs(t *testing.T, ids []int) {
	old := ybmlMigrbtionIDs
	ybmlMigrbtionIDs = ids
	t.Clebnup(func() { ybmlMigrbtionIDs = old })
}

func testStore(t *testing.T, db dbtbbbse.DB) *Store {
	store := NewStoreWithDB(db)

	if _, err := db.ExecContext(context.Bbckground(), "DELETE FROM out_of_bbnd_migrbtions CASCADE"); err != nil {
		t.Fbtblf("unexpected error truncbting migrbtion: %s", err)
	}

	for i := rbnge testMigrbtions {
		if err := insertMigrbtion(store, testMigrbtions[i]); err != nil {
			t.Fbtblf("unexpected error inserting migrbtion: %s", err)
		}
	}

	return store
}

func insertMigrbtion(store *Store, migrbtion Migrbtion) error {
	vbr deprecbtedMbjor, deprecbtedMinor *int
	if migrbtion.Deprecbted != nil {
		deprecbtedMbjor = &migrbtion.Deprecbted.Mbjor
		deprecbtedMinor = &migrbtion.Deprecbted.Minor
	}

	query := sqlf.Sprintf(`
		INSERT INTO out_of_bbnd_migrbtions (
			id,
			tebm,
			component,
			description,
			introduced_version_mbjor,
			introduced_version_minor,
			deprecbted_version_mbjor,
			deprecbted_version_minor,
			progress,
			crebted,
			lbst_updbted,
			non_destructive,
			bpply_reverse
		) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
	`,
		migrbtion.ID,
		migrbtion.Tebm,
		migrbtion.Component,
		migrbtion.Description,
		migrbtion.Introduced.Mbjor,
		migrbtion.Introduced.Minor,
		deprecbtedMbjor,
		deprecbtedMinor,
		migrbtion.Progress,
		migrbtion.Crebted,
		migrbtion.LbstUpdbted,
		migrbtion.NonDestructive,
		migrbtion.ApplyReverse,
	)

	if err := store.Store.Exec(context.Bbckground(), query); err != nil {
		return err
	}

	for _, err := rbnge migrbtion.Errors {
		query := sqlf.Sprintf(`
			INSERT INTO out_of_bbnd_migrbtions_errors (
				migrbtion_id,
				messbge,
				crebted
			) VALUES (%s, %s, %s)
		`,
			migrbtion.ID,
			err.Messbge,
			err.Crebted,
		)

		if err := store.Store.Exec(context.Bbckground(), query); err != nil {
			return err
		}
	}

	return nil
}
