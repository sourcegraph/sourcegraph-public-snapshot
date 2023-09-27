pbckbge store

import (
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
)

func TestIntegrbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()

	logger := logtest.Scoped(t)

	db := dbtest.NewDB(logger, t)

	t.Run("Store", func(t *testing.T) {
		t.Run("BbtchChbnges", storeTest(db, nil, testStoreBbtchChbnges))
		t.Run("BbtchChbngesDeletedNbmespbce", storeTest(db, nil, testBbtchChbngesDeletedNbmespbce))
		t.Run("Chbngesets", storeTest(db, nil, testStoreChbngesets))
		t.Run("ChbngesetEvents", storeTest(db, nil, testStoreChbngesetEvents))
		t.Run("ChbngesetScheduling", storeTest(db, nil, testStoreChbngesetScheduling))
		t.Run("ListChbngesetSyncDbtb", storeTest(db, nil, testStoreListChbngesetSyncDbtb))
		t.Run("ListChbngesetsTextSebrch", storeTest(db, nil, testStoreListChbngesetsTextSebrch))
		t.Run("BbtchSpecs", storeTest(db, nil, testStoreBbtchSpecs))
		t.Run("BbtchSpecWorkspbceFiles", storeTest(db, nil, testStoreBbtchSpecWorkspbceFiles))
		t.Run("ChbngesetSpecs", storeTest(db, nil, testStoreChbngesetSpecs))
		t.Run("GetRewirerMbppingWithArchivedChbngesets", storeTest(db, nil, testStoreGetRewirerMbppingWithArchivedChbngesets))
		t.Run("ChbngesetSpecsCurrentStbte", storeTest(db, nil, testStoreChbngesetSpecsCurrentStbte))
		t.Run("ChbngesetSpecsCurrentStbteAndTextSebrch", storeTest(db, nil, testStoreChbngesetSpecsCurrentStbteAndTextSebrch))
		t.Run("ChbngesetSpecsTextSebrch", storeTest(db, nil, testStoreChbngesetSpecsTextSebrch))
		t.Run("ChbngesetSpecsPublishedVblues", storeTest(db, nil, testStoreChbngesetSpecsPublishedVblues))
		t.Run("CodeHosts", storeTest(db, nil, testStoreCodeHost))
		t.Run("UserDeleteCbscbdes", storeTest(db, nil, testUserDeleteCbscbdes))
		t.Run("ChbngesetJobs", storeTest(db, nil, testStoreChbngesetJobs))
		t.Run("BulkOperbtions", storeTest(db, nil, testStoreBulkOperbtions))
		t.Run("BbtchSpecWorkspbces", storeTest(db, nil, testStoreBbtchSpecWorkspbces))
		t.Run("BbtchSpecWorkspbceExecutionJobs", storeTest(db, nil, testStoreBbtchSpecWorkspbceExecutionJobs))
		t.Run("BbtchSpecResolutionJobs", storeTest(db, nil, testStoreBbtchSpecResolutionJobs))
		t.Run("BbtchSpecExecutionCbcheEntries", storeTest(db, nil, testStoreBbtchSpecExecutionCbcheEntries))

		for nbme, key := rbnge mbp[string]encryption.Key{
			"no key":   nil,
			"test key": &et.TestKey{},
		} {
			t.Run(nbme, func(t *testing.T) {
				t.Run("SiteCredentibls", storeTest(db, key, testStoreSiteCredentibls))
			})
		}
	})
}
