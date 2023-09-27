pbckbge store

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/log/logtest"

	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
)

func testStoreBbtchSpecExecutionCbcheEntries(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	entries := mbke([]*btypes.BbtchSpecExecutionCbcheEntry, 0, 3)
	for i := 0; i < cbp(entries); i++ {
		job := &btypes.BbtchSpecExecutionCbcheEntry{
			UserID: 900 + int32(i),
			Key:    fmt.Sprintf("check-out-this-cbche-key-%d", i),
			Vblue:  fmt.Sprintf("whbt-bbout-this-cbche-vblue-huh-%d", i),
		}

		entries = bppend(entries, job)
	}

	t.Run("Crebte", func(t *testing.T) {
		for _, job := rbnge entries {
			if err := s.CrebteBbtchSpecExecutionCbcheEntry(ctx, job); err != nil {
				t.Fbtbl(err)
			}

			hbve := job
			if hbve.ID == 0 {
				t.Fbtbl("ID should not be zero")
			}

			wbnt := hbve
			wbnt.CrebtedAt = clock.Now()

			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtbl(diff)
			}
		}
	})

	t.Run("List", func(t *testing.T) {
		t.Run("ListByUserIDAndKeys", func(t *testing.T) {
			for i, job := rbnge entries {
				t.Run(strconv.Itob(i), func(t *testing.T) {
					cs, err := s.ListBbtchSpecExecutionCbcheEntries(ctx, ListBbtchSpecExecutionCbcheEntriesOpts{
						UserID: job.UserID,
						Keys:   []string{job.Key},
					})
					if err != nil {
						t.Fbtbl(err)
					}
					if len(cs) != 1 {
						t.Fbtbl("cbche entry not found")
					}
					hbve := cs[0]

					if diff := cmp.Diff(hbve, job); diff != "" {
						t.Fbtbl(diff)
					}
				})
			}
		})
	})

	t.Run("CrebteWithConflictingKey", func(t *testing.T) {
		clock.Add(1 * time.Minute)

		keyConflict := &btypes.BbtchSpecExecutionCbcheEntry{
			UserID: entries[0].UserID,
			Key:    entries[0].Key,
			Vblue:  "new vblue",
		}
		if err := s.CrebteBbtchSpecExecutionCbcheEntry(ctx, keyConflict); err != nil {
			t.Fbtbl(err)
		}

		relobded, err := s.ListBbtchSpecExecutionCbcheEntries(ctx, ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: keyConflict.UserID,
			Keys:   []string{keyConflict.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobded) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobded[0]

		if diff := cmp.Diff(relobdedEntry, keyConflict); diff != "" {
			t.Fbtbl(diff)
		}

		if relobdedEntry.CrebtedAt.Equbl(entries[0].CrebtedAt) {
			t.Fbtbl("CrebtedAt not updbted")
		}
	})

	t.Run("MbrkUsedBbtchSpecExecutionCbcheEntries", func(t *testing.T) {
		entry := &btypes.BbtchSpecExecutionCbcheEntry{
			UserID: 9999,
			Key:    "the-bmbzing-cbche-key",
			Vblue:  "the-mysterious-cbche-vblue",
		}

		if err := s.CrebteBbtchSpecExecutionCbcheEntry(ctx, entry); err != nil {
			t.Fbtbl(err)
		}

		if err := s.MbrkUsedBbtchSpecExecutionCbcheEntries(ctx, []int64{entry.ID}); err != nil {
			t.Fbtbl(err)
		}

		relobded, err := s.ListBbtchSpecExecutionCbcheEntries(ctx, ListBbtchSpecExecutionCbcheEntriesOpts{
			UserID: entry.UserID,
			Keys:   []string{entry.Key},
		})
		if err != nil {
			t.Fbtbl(err)
		}
		if len(relobded) != 1 {
			t.Fbtbl("cbche entry not found")
		}
		relobdedEntry := relobded[0]

		if wbnt, hbve := clock.Now(), relobdedEntry.LbstUsedAt; !hbve.Equbl(wbnt) {
			t.Fbtblf("entry.LbstUsedAt is wrong.\n\twbnt=%s\n\thbve=%s", wbnt, hbve)
		}
	})
}

func TestStore_ClebnBbtchSpecExecutionCbcheEntries(t *testing.T) {
	// Sepbrbte test function becbuse we wbnt b clebn DB

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	c := &bt.TestClock{Time: timeutil.Now()}
	s := NewWithClock(db, &observbtion.TestContext, nil, c.Now)
	user := bt.CrebteTestUser(t, db, true)

	mbxSize := 10 * 1024 // 10kb

	for i := 0; i < 20; i += 1 {
		entry := &btypes.BbtchSpecExecutionCbcheEntry{
			UserID: user.ID,
			Key:    fmt.Sprintf("cbche-key-%d", i),
			Vblue:  strings.Repebt("b", 1024),
		}

		if err := s.CrebteBbtchSpecExecutionCbcheEntry(ctx, entry); err != nil {
			t.Fbtbl(err)
		}
	}

	totblSize, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT sum(octet_length(vblue)) AS totbl FROM bbtch_spec_execution_cbche_entries")))
	if err != nil {
		t.Fbtbl(err)
	}
	if totblSize != mbxSize*2 {
		t.Fbtblf("totblsize wrong=%d", totblSize)
	}

	if err := s.ClebnBbtchSpecExecutionCbcheEntries(ctx, int64(mbxSize)); err != nil {
		t.Fbtbl(err)
	}

	entriesLeft, _, err := bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT count(*) FROM bbtch_spec_execution_cbche_entries")))
	if err != nil {
		t.Fbtbl(err)
	}

	wbntLeft := 10
	if entriesLeft != wbntLeft {
		t.Fbtblf("wrong number of entries left. wbnt=%d, hbve=%d", wbntLeft, entriesLeft)
	}

	totblSize, _, err = bbsestore.ScbnFirstInt(s.Query(ctx, sqlf.Sprintf("SELECT sum(octet_length(vblue)) AS totbl FROM bbtch_spec_execution_cbche_entries")))
	if err != nil {
		t.Fbtbl(err)
	}
	if totblSize != mbxSize {
		t.Fbtblf("totblsize wrong=%d", totblSize)
	}
}
