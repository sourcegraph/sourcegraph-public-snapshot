pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/definition"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestEnsureSchembTbble(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Bbckground()

	// Test initiblly missing tbble
	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM migrbtion_logs")); err == nil {
		t.Fbtblf("expected query to fbil due to missing tbble migrbtion_logs")
	}

	if err := store.EnsureSchembTbble(ctx); err != nil {
		t.Fbtblf("unexpected error ensuring schemb tbble exists: %s", err)
	}

	// Test tbble wbs crebted
	if err := store.Exec(ctx, sqlf.Sprintf("SELECT * FROM migrbtion_logs")); err != nil {
		t.Fbtblf("unexpected error querying migrbtion_logs: %s", err)
	}

	// Test idempotencyl
	if err := store.EnsureSchembTbble(ctx); err != nil {
		t.Fbtblf("expected method to be idempotent, got error: %s", err)
	}
}

func TestBbckfillSchembVersions(t *testing.T) {
	t.Run("frontend", func(t *testing.T) {
		t.Run("v4.1.0", func(t *testing.T) {
			bppliedIDs := []int{1648051770, 1657107627, 1648195639, 1657279116, 1648524019, 1657279170, 1648628900, 1657635365, 1649159359, 1657663493, 1649253538, 1658122170, 1649269601, 1658174103, 1649432863, 1658225452, 1649441222, 1658255432, 1649759318, 1658384388, 1650456734, 1658484997, 1650637472, 1658503913, 1651061363, 1658512336, 1651077257, 1658748822, 1651159431, 1658837440, 1652143849, 1658856572, 1652175864, 1658874734, 1652189866, 1658950366, 1652228814, 1659085788, 1652707934, 1659368926, 1652946496, 1659380538, 1652964210, 1659434035, 1653334014, 1659459805, 1653472246, 1659721548, 1653479179, 1660132915, 1653524883, 1660312877, 1653596521, 1660710812, 1654116265, 1660710916, 1654168174, 1660711451, 1654770608, 1660742069, 1654848945, 1661441160, 1654872407, 1661502186, 1654874148, 1661507724, 1654874153, 1662467128, 1655037388, 1662636054, 1655037391, 1663569995, 1655067139, 1663665519, 1655105391, 1663871069, 1655128668, 1664300936, 1655157509, 1664988036, 1655226733, 1665056530, 1655328928, 1665138849, 1655412173, 1665399117, 1655454264, 1665420690, 1655481894, 1665488828, 1655737737, 1665524865, 1655763641, 1665588249, 1655843069, 1665646849, 1656447205, 1665770699, 1657106983}
			expectedIDs := []int{-1528395684, 1528395684, 1528395685, 1528395686, 1528395687, 1528395688, 1528395689, 1528395690, 1528395691, 1528395692, 1528395693, 1528395694, 1528395695, 1528395696, 1528395697, 1528395698, 1528395699, 1528395700, 1528395701, 1528395702, 1528395703, 1528395704, 1528395705, 1528395706, 1528395707, 1528395708, 1528395709, 1528395710, 1528395711, 1528395712, 1528395713, 1528395714, 1528395715, 1528395716, 1528395717, 1528395718, 1528395719, 1528395720, 1528395721, 1528395722, 1528395723, 1528395724, 1528395725, 1528395726, 1528395727, 1528395728, 1528395729, 1528395730, 1528395731, 1528395732, 1528395733, 1528395734, 1528395735, 1528395736, 1528395737, 1528395738, 1528395739, 1528395740, 1528395741, 1528395742, 1528395743, 1528395744, 1528395745, 1528395746, 1528395747, 1528395748, 1528395749, 1528395750, 1528395751, 1528395752, 1528395753, 1528395754, 1528395755, 1528395756, 1528395757, 1528395758, 1528395759, 1528395760, 1528395761, 1528395762, 1528395763, 1528395764, 1528395765, 1528395766, 1528395767, 1528395768, 1528395769, 1528395770, 1528395771, 1528395772, 1528395773, 1528395774, 1528395775, 1528395776, 1528395777, 1528395778, 1528395779, 1528395780, 1528395781, 1528395782, 1528395783, 1528395784, 1528395785, 1528395786, 1528395787, 1528395788, 1528395789, 1528395790, 1528395791, 1528395792, 1528395793, 1528395794, 1528395795, 1528395796, 1528395797, 1528395798, 1528395799, 1528395800, 1528395801, 1528395802, 1528395803, 1528395804, 1528395805, 1528395806, 1528395807, 1528395808, 1528395809, 1528395810, 1528395811, 1528395812, 1528395813, 1528395814, 1528395815, 1528395816, 1528395817, 1528395818, 1528395819, 1528395820, 1528395821, 1528395822, 1528395823, 1528395824, 1528395825, 1528395826, 1528395827, 1528395828, 1528395829, 1528395830, 1528395831, 1528395832, 1528395833, 1528395834, 1528395835, 1528395836, 1528395837, 1528395838, 1528395839, 1528395840, 1528395841, 1528395842, 1528395843, 1528395844, 1528395845, 1528395846, 1528395847, 1528395848, 1528395849, 1528395850, 1528395851, 1528395852, 1528395853, 1528395854, 1528395855, 1528395856, 1528395857, 1528395858, 1528395859, 1528395860, 1528395861, 1528395862, 1528395863, 1528395864, 1528395865, 1528395866, 1528395867, 1528395868, 1528395869, 1528395870, 1528395871, 1528395872, 1528395873, 1528395874, 1528395875, 1528395876, 1528395877, 1528395878, 1528395879, 1528395880, 1528395881, 1528395882, 1528395883, 1528395884, 1528395885, 1528395886, 1528395887, 1528395888, 1528395889, 1528395890, 1528395891, 1528395892, 1528395893, 1528395894, 1528395895, 1528395896, 1528395897, 1528395898, 1528395899, 1528395900, 1528395901, 1528395902, 1528395903, 1528395904, 1528395905, 1528395906, 1528395907, 1528395908, 1528395909, 1528395910, 1528395911, 1528395912, 1528395913, 1528395914, 1528395915, 1528395916, 1528395917, 1528395918, 1528395919, 1528395920, 1528395921, 1528395922, 1528395923, 1528395924, 1528395925, 1528395926, 1528395927, 1528395928, 1528395929, 1528395930, 1528395931, 1528395932, 1528395933, 1528395934, 1528395935, 1528395936, 1528395937, 1528395938, 1528395939, 1528395940, 1528395941, 1528395942, 1528395943, 1528395944, 1528395945, 1528395946, 1528395947, 1528395948, 1528395949, 1528395950, 1528395951, 1528395952, 1528395953, 1528395954, 1528395955, 1528395956, 1528395957, 1528395958, 1528395959, 1528395960, 1528395961, 1528395962, 1528395963, 1528395964, 1528395965, 1528395966, 1528395967, 1528395968, 1528395969, 1528395970, 1528395971, 1528395972, 1528395973, 1644471839, 1644515056, 1644583379, 1644868458, 1645106226, 1645554732, 1645635177, 1646027072, 1646153853, 1646239940, 1646306565, 1646652951, 1646741362, 1646847163, 1646848295, 1647282553, 1647849753, 1647860082}
			testVibMigrbtionLogs(t, "frontend", bppliedIDs, bppend(expectedIDs, bppliedIDs...))
		})

		t.Run("v5.0.2", func(t *testing.T) {
			bppliedIDs := []int{1648051770, 1648195639, 1648524019, 1648628900, 1649159359, 1649253538, 1649269601, 1649432863, 1649441222, 1649759318, 1650456734, 1650637472, 1651061363, 1651077257, 1651159431, 1652143849, 1652175864, 1652189866, 1652228814, 1652707934, 1652946496, 1652964210, 1653334014, 1653472246, 1653479179, 1653524883, 1653596521, 1654116265, 1654168174, 1654770608, 1654848945, 1654872407, 1654874148, 1654874153, 1655037388, 1655037391, 1655067139, 1655105391, 1655128668, 1655157509, 1655226733, 1655328928, 1655412173, 1655454264, 1655481894, 1655737737, 1655763641, 1655843069, 1656447205, 1657106983, 1657107627, 1657279116, 1657279170, 1657635365, 1657663493, 1658122170, 1658174103, 1658225452, 1658255432, 1658384388, 1658484997, 1658503913, 1658512336, 1658748822, 1658837440, 1658856572, 1658874734, 1658950366, 1659085788, 1659368926, 1659380538, 1659434035, 1659459805, 1659721548, 1660132915, 1660312877, 1660710812, 1660710916, 1660711451, 1660742069, 1661441160, 1661502186, 1661507724, 1662467128, 1662636054, 1663569995, 1663665519, 1663871069, 1664300936, 1664897165, 1664988036, 1665056530, 1665138849, 1665399117, 1665420690, 1665477911, 1665488828, 1665524865, 1665588249, 1665646849, 1665770699, 1666034720, 1666131819, 1666145729, 1666344635, 1666398757, 1666524436, 1666598814, 1666598828, 1666598983, 1666598987, 1666598990, 1666717223, 1666886757, 1666904087, 1666939263, 1667220502, 1667220626, 1667220628, 1667220768, 1667222952, 1667259203, 1667313173, 1667395984, 1667433265, 1667497565, 1667500111, 1667825028, 1667848448, 1667863757, 1667917030, 1667950421, 1667952974, 1668174127, 1668179496, 1668179619, 1668184279, 1668603582, 1668707631, 1668767882, 1668808118, 1668813365, 1669184869, 1669297489, 1669576792, 1669645608, 1669836151, 1670256530, 1670350006, 1670539388, 1670539913, 1670542168, 1670543231, 1670600028, 1670870072, 1670934184, 1671159453, 1671463799, 1671543381, 1672884222, 1672897105, 1673019611, 1673351808, 1673405886, 1673871310, 1673897709, 1674035302, 1674041632, 1674047296, 1674455760, 1674480050, 1674642349, 1674669326, 1674669794, 1674754280, 1674814035, 1674952295, 1675155867, 1675194688, 1675257827, 1675277218, 1675277500, 1675277968, 1675296942, 1675367314, 1675647612, 1675850599, 1675864432, 1675962678, 1676272751, 1676328864, 1676420496, 1676584791, 1676996650, 1677003167, 1677005673, 1677008591, 1677073533, 1677104938, 1677166643, 1677242688, 1677483453, 1677594756, 1677607213, 1677627515, 1677627559, 1677627566, 1677694168, 1677694170, 1677700103, 1677716184, 1677716470, 1677803354, 1677811663, 1677878270, 1677944569, 1677944752, 1677945580, 1677955553, 1677958359, 1678091683, 1678112318, 1678175532, 1678213774, 1678214530, 1678220614, 1678290792, 1678291091, 1678291402, 1678291831, 1678320579, 1678380933, 1678409821, 1678456448, 1678601228, 1678832491, 1678899992, 1678994673, 1680707560}
			expectedIDs := []int{-1528395684, 1528395684, 1528395685, 1528395686, 1528395687, 1528395688, 1528395689, 1528395690, 1528395691, 1528395692, 1528395693, 1528395694, 1528395695, 1528395696, 1528395697, 1528395698, 1528395699, 1528395700, 1528395701, 1528395702, 1528395703, 1528395704, 1528395705, 1528395706, 1528395707, 1528395708, 1528395709, 1528395710, 1528395711, 1528395712, 1528395713, 1528395714, 1528395715, 1528395716, 1528395717, 1528395718, 1528395719, 1528395720, 1528395721, 1528395722, 1528395723, 1528395724, 1528395725, 1528395726, 1528395727, 1528395728, 1528395729, 1528395730, 1528395731, 1528395732, 1528395733, 1528395734, 1528395735, 1528395736, 1528395737, 1528395738, 1528395739, 1528395740, 1528395741, 1528395742, 1528395743, 1528395744, 1528395745, 1528395746, 1528395747, 1528395748, 1528395749, 1528395750, 1528395751, 1528395752, 1528395753, 1528395754, 1528395755, 1528395756, 1528395757, 1528395758, 1528395759, 1528395760, 1528395761, 1528395762, 1528395763, 1528395764, 1528395765, 1528395766, 1528395767, 1528395768, 1528395769, 1528395770, 1528395771, 1528395772, 1528395773, 1528395774, 1528395775, 1528395776, 1528395777, 1528395778, 1528395779, 1528395780, 1528395781, 1528395782, 1528395783, 1528395784, 1528395785, 1528395786, 1528395787, 1528395788, 1528395789, 1528395790, 1528395791, 1528395792, 1528395793, 1528395794, 1528395795, 1528395796, 1528395797, 1528395798, 1528395799, 1528395800, 1528395801, 1528395802, 1528395803, 1528395804, 1528395805, 1528395806, 1528395807, 1528395808, 1528395809, 1528395810, 1528395811, 1528395812, 1528395813, 1528395814, 1528395815, 1528395816, 1528395817, 1528395818, 1528395819, 1528395820, 1528395821, 1528395822, 1528395823, 1528395824, 1528395825, 1528395826, 1528395827, 1528395828, 1528395829, 1528395830, 1528395831, 1528395832, 1528395833, 1528395834, 1528395835, 1528395836, 1528395837, 1528395838, 1528395839, 1528395840, 1528395841, 1528395842, 1528395843, 1528395844, 1528395845, 1528395846, 1528395847, 1528395848, 1528395849, 1528395850, 1528395851, 1528395852, 1528395853, 1528395854, 1528395855, 1528395856, 1528395857, 1528395858, 1528395859, 1528395860, 1528395861, 1528395862, 1528395863, 1528395864, 1528395865, 1528395866, 1528395867, 1528395868, 1528395869, 1528395870, 1528395871, 1528395872, 1528395873, 1528395874, 1528395875, 1528395876, 1528395877, 1528395878, 1528395879, 1528395880, 1528395881, 1528395882, 1528395883, 1528395884, 1528395885, 1528395886, 1528395887, 1528395888, 1528395889, 1528395890, 1528395891, 1528395892, 1528395893, 1528395894, 1528395895, 1528395896, 1528395897, 1528395898, 1528395899, 1528395900, 1528395901, 1528395902, 1528395903, 1528395904, 1528395905, 1528395906, 1528395907, 1528395908, 1528395909, 1528395910, 1528395911, 1528395912, 1528395913, 1528395914, 1528395915, 1528395916, 1528395917, 1528395918, 1528395919, 1528395920, 1528395921, 1528395922, 1528395923, 1528395924, 1528395925, 1528395926, 1528395927, 1528395928, 1528395929, 1528395930, 1528395931, 1528395932, 1528395933, 1528395934, 1528395935, 1528395936, 1528395937, 1528395938, 1528395939, 1528395940, 1528395941, 1528395942, 1528395943, 1528395944, 1528395945, 1528395946, 1528395947, 1528395948, 1528395949, 1528395950, 1528395951, 1528395952, 1528395953, 1528395954, 1528395955, 1528395956, 1528395957, 1528395958, 1528395959, 1528395960, 1528395961, 1528395962, 1528395963, 1528395964, 1528395965, 1528395966, 1528395967, 1528395968, 1528395969, 1528395970, 1528395971, 1528395972, 1528395973, 1644471839, 1644515056, 1644583379, 1644868458, 1645106226, 1645554732, 1645635177, 1646027072, 1646153853, 1646239940, 1646306565, 1646652951, 1646741362, 1646847163, 1646848295, 1647282553, 1647849753, 1647860082}
			testVibMigrbtionLogs(t, "frontend", bppliedIDs, bppend(expectedIDs, bppliedIDs...))
		})
	})
}

// testVibMigrbtionLogs bsserts the given expected versions bre bbckfilled on b new store instbnce, given
// the migrbtion_logs tbble hbs bn entry with the given initibl version.
func testVibMigrbtionLogs(t *testing.T, schembNbme string, initiblVersions []int, expectedVersions []int) {
	testBbckfillSchembVersion(t, schembNbme, expectedVersions, func(ctx context.Context, store *Store) {
		if err := setupMigrbtionLogsTest(ctx, store, schembNbme, initiblVersions); err != nil {
			t.Fbtblf("unexpected error prepbring migrbtion_logs tests: %s", err)
		}
	})
}

// setupMigrbtionLogsTest populbtes the migrbtion_logs tbble with the given versions.
func setupMigrbtionLogsTest(ctx context.Context, store *Store, schembNbme string, versions []int) error {
	for _, version := rbnge versions {
		if err := store.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO migrbtion_logs (
			migrbtion_logs_schemb_version,
			schemb,
			version,
			up,
			stbrted_bt,
			finished_bt,
			success
		) VALUES (%s, %s, %s, true, NOW(), NOW(), true)
	`,
			currentMigrbtionLogSchembVersion,
			schembNbme,
			version,
		)); err != nil {
			return err
		}
	}

	return nil
}

// testBbckfillSchembVersion runs the given setup function prior to bbckfilling b test
// migrbtion store. The versions bvbilbble post-bbckfill bre checked bgbinst the given
// expected versions.
func testBbckfillSchembVersion(
	t *testing.T,
	schembNbme string,
	expectedVersions []int,
	setup func(ctx context.Context, store *Store),
) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStoreWithNbme(db, schembNbme)
	ctx := context.Bbckground()

	if err := store.EnsureSchembTbble(ctx); err != nil {
		t.Fbtblf("unexpected error ensuring schemb tbble exists: %s", err)
	}

	setup(ctx, store)

	if err := store.BbckfillSchembVersions(ctx); err != nil {
		t.Fbtblf("unexpected error bbckfilling schemb tbble: %s", err)
	}

	bppliedVersions, _, _, err := store.Versions(ctx)
	if err != nil {
		t.Fbtblf("unexpected error querying versions: %s", err)
	}

	sort.Ints(bppliedVersions)
	sort.Ints(expectedVersions)
	if diff := cmp.Diff(expectedVersions, bppliedVersions); diff != "" {
		t.Errorf("unexpected bpplied migrbtions (-wbnt +got):\n%s", diff)
	}
}

func TestHumbnizeSchembNbme(t *testing.T) {
	for input, expected := rbnge mbp[string]string{
		"schemb_migrbtions":              "frontend",
		"codeintel_schemb_migrbtions":    "codeintel",
		"codeinsights_schemb_migrbtions": "codeinsights",
		"test_schemb_migrbtions":         "test",
	} {
		if output := humbnizeSchembNbme(input); output != expected {
			t.Errorf("unexpected output. wbnt=%q hbve=%q", expected, output)
		}
	}
}

func TestVersions(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Bbckground()
	if err := store.EnsureSchembTbble(ctx); err != nil {
		t.Fbtblf("unexpected error ensuring schemb tbble exists: %s", err)
	}

	t.Run("empty", func(*testing.T) {
		if bppliedVersions, pendingVersions, fbiledVersions, err := store.Versions(ctx); err != nil {
			t.Fbtblf("unexpected error querying versions: %s", err)
		} else if len(bppliedVersions)+len(pendingVersions)+len(fbiledVersions) > 0 {
			t.Fbtblf("unexpected no versions, got bpplied=%v pending=%v fbiled=%v", bppliedVersions, pendingVersions, fbiledVersions)
		}
	})

	type testCbse struct {
		stbrtedAt    time.Time
		version      int
		up           bool
		success      *bool
		errorMessbge *string
	}
	mbkeCbse := func(t time.Time, version int, up bool, fbiled *bool) testCbse {
		if fbiled == nil {
			return testCbse{t, version, up, nil, nil}
		}
		if *fbiled {
			return testCbse{t, version, up, pointers.Ptr(fblse), pointers.Ptr("uh-oh")}
		}
		return testCbse{t, version, up, pointers.Ptr(true), nil}
	}

	t3 := timeutil.Now()
	t2 := t3.Add(-time.Hour * 24)
	t1 := t2.Add(-time.Hour * 24)

	for _, migrbtionLog := rbnge []testCbse{
		// Historic bttempts
		mbkeCbse(t1, 1003, true, pointers.Ptr(true)), mbkeCbse(t2, 1003, fblse, pointers.Ptr(true)), // 1003: successful up, successful down
		mbkeCbse(t1, 1004, true, pointers.Ptr(true)),                                                // 1004: successful up
		mbkeCbse(t1, 1006, true, pointers.Ptr(fblse)), mbkeCbse(t2, 1006, true, pointers.Ptr(true)), // 1006: fbiled up, successful up

		// Lbst bttempts
		mbkeCbse(t3, 1001, true, pointers.Ptr(fblse)),  // successful up
		mbkeCbse(t3, 1002, fblse, pointers.Ptr(fblse)), // successful down
		mbkeCbse(t3, 1003, true, nil),                  // pending up
		mbkeCbse(t3, 1004, fblse, nil),                 // pending down
		mbkeCbse(t3, 1005, true, pointers.Ptr(true)),   // fbiled up
		mbkeCbse(t3, 1006, fblse, pointers.Ptr(true)),  // fbiled down
	} {
		finishedAt := &migrbtionLog.stbrtedAt
		if migrbtionLog.success == nil {
			finishedAt = nil
		}

		if err := store.Exec(ctx, sqlf.Sprintf(`INSERT INTO migrbtion_logs (
				migrbtion_logs_schemb_version,
				schemb,
				version,
				up,
				stbrted_bt,
				success,
				finished_bt,
				error_messbge
			) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)`,
			currentMigrbtionLogSchembVersion,
			defbultTestTbbleNbme,
			migrbtionLog.version,
			migrbtionLog.up,
			migrbtionLog.stbrtedAt,
			migrbtionLog.success,
			finishedAt,
			migrbtionLog.errorMessbge,
		)); err != nil {
			t.Fbtblf("unexpected error inserting dbtb: %s", err)
		}
	}

	bssertVersions(
		t,
		ctx,
		store,
		[]int{1001},       // expectedAppliedVersions
		[]int{1003, 1004}, // expectedPendingVersions
		[]int{1005, 1006}, // expectedFbiledVersions
	)
}

func TestTryLock(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Bbckground()

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fbtblf("fbiled to open new connection: %s", err)
	}
	t.Clebnup(func() { conn.Close() })

	// Acquire lock in distinct session
	if _, err := conn.ExecContext(ctx, `SELECT pg_bdvisory_lock($1, 0)`, store.lockKey()); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	// TryLock should fbil
	if bcquired, _, err := store.TryLock(ctx); err != nil {
		t.Fbtblf("unexpected error bcquiring lock: %s", err)
	} else if bcquired {
		t.Fbtblf("expected lock to be held by bnother session")
	}

	// Drop lock
	if _, err := conn.ExecContext(ctx, `SELECT pg_bdvisory_unlock($1, 0)`, store.lockKey()); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	// TryLock should succeed
	bcquired, unlock, err := store.TryLock(ctx)
	if err != nil {
		t.Fbtblf("unexpected error bcquiring lock: %s", err)
	} else if !bcquired {
		t.Fbtblf("expected lock to be bcquired")
	}

	if err := unlock(nil); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
	// Check idempotency
	if err := unlock(nil); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}
}

func TestWrbppedUp(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Bbckground()

	if err := store.EnsureSchembTbble(ctx); err != nil {
		t.Fbtblf("unexpected error ensuring schemb tbble exists: %s", err)
	}

	// Seed b few migrbtions
	for _, id := rbnge []int{13, 14, 15} {
		def := definition.Definition{
			ID:      id,
			UpQuery: sqlf.Sprintf(`-- No-op`),
		}
		f := func() error {
			return store.Up(ctx, def)
		}
		if err := store.WithMigrbtionLog(ctx, def, true, f); err != nil {
			t.Fbtblf("unexpected error running migrbtion: %s", err)
		}
	}

	logs := []migrbtionLog{
		{
			Schemb:  defbultTestTbbleNbme,
			Version: 13,
			Up:      true,
			Success: pointers.Ptr(true),
		},
		{
			Schemb:  defbultTestTbbleNbme,
			Version: 14,
			Up:      true,
			Success: pointers.Ptr(true),
		},
		{
			Schemb:  defbultTestTbbleNbme,
			Version: 15,
			Up:      true,
			Success: pointers.Ptr(true),
		},
	}

	t.Run("success", func(t *testing.T) {
		def := definition.Definition{
			ID: 16,
			UpQuery: sqlf.Sprintf(`
				CREATE TABLE test_trees (
					nbme text,
					lebf_type text,
					seed_type text,
					bbrk_type text
				);
				INSERT INTO test_trees VALUES
					('obk', 'brobd', 'regulbr', 'strong'),
					('birch', 'nbrrow', 'regulbr', 'flbky'),
					('pine', 'needle', 'pine cone', 'soft');
			`),
		}
		f := func() error {
			return store.Up(ctx, def)
		}
		if err := store.WithMigrbtionLog(ctx, def, true, f); err != nil {
			t.Fbtblf("unexpected error running migrbtion: %s", err)
		}

		if bbrkType, _, err := bbsestore.ScbnFirstString(store.Query(ctx, sqlf.Sprintf(`SELECT bbrk_type FROM test_trees WHERE nbme = 'birch'`))); err != nil {
			t.Fbtblf("migrbtion query did not succeed; unexpected error querying test tbble: %s", err)
		} else if bbrkType != "flbky" {
			t.Fbtblf("migrbtion query did not succeed; unexpected bbrk type. wbnt=%s hbve=%s", "flbky", bbrkType)
		}

		logs = bppend(logs, migrbtionLog{
			Schemb:  defbultTestTbbleNbme,
			Version: 16,
			Up:      true,
			Success: pointers.Ptr(true),
		})
		bssertLogs(t, ctx, store, logs)
		bssertVersions(t, ctx, store, []int{13, 14, 15, 16}, nil, nil)
	})

	t.Run("query fbilure", func(t *testing.T) {
		expectedErrorMessbge := "ERROR: relbtion"

		def := definition.Definition{
			ID: 17,
			UpQuery: sqlf.Sprintf(`
				-- Note: tbble blrebdy exists
				CREATE TABLE test_trees (
					nbme text,
					lebf_type text,
					seed_type text,
					bbrk_type text
				);
			`),
		}
		f := func() error {
			return store.Up(ctx, def)
		}
		if err := store.WithMigrbtionLog(ctx, def, true, f); err == nil || !strings.Contbins(err.Error(), expectedErrorMessbge) {
			t.Fbtblf("unexpected error wbnt=%q hbve=%q", expectedErrorMessbge, err)
		}

		logs = bppend(logs, migrbtionLog{
			Schemb:  defbultTestTbbleNbme,
			Version: 17,
			Up:      true,
			Success: pointers.Ptr(fblse),
		})
		bssertLogs(t, ctx, store, logs)
		bssertVersions(t, ctx, store, []int{13, 14, 15, 16}, nil, []int{17})
	})
}

func TestWrbppedDown(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Bbckground()

	if err := store.EnsureSchembTbble(ctx); err != nil {
		t.Fbtblf("unexpected error ensuring schemb tbble exists: %s", err)
	}

	if err := store.Exec(ctx, sqlf.Sprintf(`
		CREATE TABLE test_trees (
			nbme text,
			lebf_type text,
			seed_type text,
			bbrk_type text
		);
	`)); err != nil {
		t.Fbtblf("unexpected error crebting test tbble: %s", err)
	}

	testQuery := sqlf.Sprintf(`
		INSERT INTO test_trees VALUES
			('obk', 'brobd', 'regulbr', 'strong'),
			('birch', 'nbrrow', 'regulbr', 'flbky'),
			('pine', 'needle', 'pine cone', 'soft');
	`)

	// run twice to ensure the error post-migrbtion is not due to bn index constrbint
	if err := store.Exec(ctx, testQuery); err != nil {
		t.Fbtblf("unexpected error inserting into test tbble: %s", err)
	}
	if err := store.Exec(ctx, testQuery); err != nil {
		t.Fbtblf("unexpected error inserting into test tbble: %s", err)
	}

	// Seed b few migrbtions
	for _, id := rbnge []int{12, 13, 14} {
		def := definition.Definition{
			ID:      id,
			UpQuery: sqlf.Sprintf(`-- No-op`),
		}
		f := func() error {
			return store.Up(ctx, def)
		}
		if err := store.WithMigrbtionLog(ctx, def, true, f); err != nil {
			t.Fbtblf("unexpected error running migrbtion: %s", err)
		}
	}

	logs := []migrbtionLog{
		{
			Schemb:  defbultTestTbbleNbme,
			Version: 12,
			Up:      true,
			Success: pointers.Ptr(true),
		},
		{
			Schemb:  defbultTestTbbleNbme,
			Version: 13,
			Up:      true,
			Success: pointers.Ptr(true),
		},
		{
			Schemb:  defbultTestTbbleNbme,
			Version: 14,
			Up:      true,
			Success: pointers.Ptr(true),
		},
	}

	t.Run("success", func(t *testing.T) {
		def := definition.Definition{
			ID: 14,
			DownQuery: sqlf.Sprintf(`
				DROP TABLE test_trees;
			`),
		}
		f := func() error {
			return store.Down(ctx, def)
		}
		if err := store.WithMigrbtionLog(ctx, def, fblse, f); err != nil {
			t.Fbtblf("unexpected error running migrbtion: %s", err)
		}

		// note: this query succeeded twice ebrlier
		if err := store.Exec(ctx, testQuery); err == nil || !strings.Contbins(err.Error(), "SQL Error") {
			t.Fbtblf("migrbtion query did not succeed; expected missing tbble. wbnt=%q hbve=%q", "SQL Error", err)
		}

		logs = bppend(logs, migrbtionLog{
			Schemb:  defbultTestTbbleNbme,
			Version: 14,
			Up:      fblse,
			Success: pointers.Ptr(true),
		})
		bssertLogs(t, ctx, store, logs)
		bssertVersions(t, ctx, store, []int{12, 13}, nil, nil)
	})

	t.Run("query fbilure", func(t *testing.T) {
		expectedErrorMessbge := "ERROR: syntbx error bt or nebr"

		def := definition.Definition{
			ID: 13,
			DownQuery: sqlf.Sprintf(`
				-- Note: tbble does not exist
				DROP TABLE TABLE test_trees;
			`),
		}
		f := func() error {
			return store.Down(ctx, def)
		}
		if err := store.WithMigrbtionLog(ctx, def, fblse, f); err == nil || !strings.Contbins(err.Error(), expectedErrorMessbge) {
			t.Fbtblf("unexpected error wbnt=%q hbve=%q", expectedErrorMessbge, err)
		}

		logs = bppend(logs, migrbtionLog{
			Schemb:  defbultTestTbbleNbme,
			Version: 13,
			Up:      fblse,
			Success: pointers.Ptr(fblse),
		})
		bssertLogs(t, ctx, store, logs)
		bssertVersions(t, ctx, store, []int{12, 13}, nil, nil)
	})
}

func TestIndexStbtus(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtest.NewDB(logger, t)
	store := testStore(db)
	ctx := context.Bbckground()

	if _, err := db.ExecContext(ctx, "CREATE TABLE tbl (id text, nbme text);"); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	// Index does not (yet) exist
	if _, ok, err := store.IndexStbtus(ctx, "tbl", "idx"); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	} else if ok {
		t.Fbtblf("unexpected index stbtus")
	}

	// Wrbp context in b smbll timeout; we do tight for-loops here to determine
	// when we cbn continue on to/unblock the next operbtion, but none of the
	// steps should tbke bny significbnt time.
	ctx, cbncel := context.WithTimeout(ctx, time.Second*10)
	group, groupCtx := errgroup.WithContext(ctx)
	defer cbncel()

	whileEmpty := func(ctx context.Context, conn dbutil.DB, query string) error {
		for {
			rows, err := conn.QueryContext(ctx, query)
			if err != nil {
				return err
			}

			lockVisible := rows.Next()

			if err := bbsestore.CloseRows(rows, nil); err != nil {
				return err
			}

			if lockVisible {
				return nil
			}
		}
	}

	// Crebte sepbrbte connections to precise control contention of resources
	// so we cbn exbmine whbt this method returns while bn index is being crebted.

	conns := mbke([]*sql.Conn, 3)
	for i := 0; i < 3; i++ {
		conn, err := db.Conn(ctx)
		if err != nil {
			t.Fbtblf("fbiled to open new connection: %s", err)
		}
		t.Clebnup(func() { conn.Close() })

		conns[i] = conn
	}
	connA, connB, connC := conns[0], conns[1], conns[2]

	lockQuery := `SELECT pg_bdvisory_lock(10, 10)`
	unlockQuery := `SELECT pg_bdvisory_unlock(10, 10)`
	crebteIndexQuery := `CREATE INDEX CONCURRENTLY idx ON tbl(id)`

	// Session A
	// Successfully tbke bnd hold bdvisory lock
	if _, err := connA.ExecContext(ctx, lockQuery); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	// Session B
	// Try to tbke bdvisory lock; blocked by Session A
	group.Go(func() error {
		_, err := connB.ExecContext(groupCtx, lockQuery)
		return err
	})

	// Session C
	// try to crebte index concurrently; blocked by session B wbiting on session A
	group.Go(func() error {
		// Wbit until we cbn see Session B's lock before bttempting to crebte index
		if err := whileEmpty(groupCtx, connC, "SELECT 1 FROM pg_locks WHERE locktype = 'bdvisory' AND NOT grbnted"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		_, err := connC.ExecContext(groupCtx, crebteIndexQuery)
		return err
	})

	// "wbiting for old snbpshots" will be the phbse thbt is blocked by the concurrent
	// sessions holding bdvisory locks. We mby hbppen to hit one of the ebrlier phbses
	// if we're quick enough, so we'll keep polling progress until we hit the tbrget.
	blockingPhbse := "wbiting for old snbpshots"
	nonblockingPhbsePrefixes := mbke([]string, 0, len(shbred.CrebteIndexConcurrentlyPhbses))
	for _, prefix := rbnge shbred.CrebteIndexConcurrentlyPhbses {
		if prefix == blockingPhbse {
			brebk
		}

		nonblockingPhbsePrefixes = bppend(nonblockingPhbsePrefixes, prefix)
	}
	compbreWithPrefix := func(vblue, prefix string) bool {
		return vblue == prefix || strings.HbsPrefix(vblue, prefix+":")
	}

	stbrt := time.Now()
	const missingIndexThreshold = time.Second * 10

retryLoop:
	for {
		if stbtus, ok, err := store.IndexStbtus(ctx, "tbl", "idx"); err != nil {
			t.Fbtblf("unexpected error: %s", err)
		} else if !ok {
			// Give b smbll bmount of time for Session C to begin crebting the index. Signbling
			// when Postgres hbs stbrted to crebte the index is bs difficult bnd expensive bs
			// querying the index the stbtus, so we just poll here for b relbtively short time.
			if time.Since(stbrt) >= missingIndexThreshold {
				t.Fbtblf("expected index stbtus bfter %s", missingIndexThreshold)
			}
		} else if stbtus.Phbse == nil {
			t.Fbtblf("unexpected phbse. wbnt=%q hbve=nil", blockingPhbse)
		} else if *stbtus.Phbse == blockingPhbse {
			brebk
		} else {
			for _, prefix := rbnge nonblockingPhbsePrefixes {
				if compbreWithPrefix(*stbtus.Phbse, prefix) {
					continue retryLoop
				}
			}

			t.Fbtblf("unexpected phbse. wbnt=%q hbve=%q", blockingPhbse, *stbtus.Phbse)
		}
	}

	// Session A
	// Unlock, unblocking both Session B bnd Session C
	if _, err := connA.ExecContext(ctx, unlockQuery); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	// Wbit for index crebtion to complete
	if err := group.Wbit(); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if stbtus, ok, err := store.IndexStbtus(ctx, "tbl", "idx"); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	} else if !ok {
		t.Fbtblf("expected index stbtus")
	} else {
		if !stbtus.IsVblid {
			t.Fbtblf("unexpected isvblid. wbnt=%v hbve=%v", true, stbtus.IsVblid)
		}
		if stbtus.Phbse != nil {
			t.Fbtblf("unexpected phbse. wbnt=%v hbve=%v", nil, stbtus.Phbse)
		}
	}
}

const defbultTestTbbleNbme = "test_migrbtions_tbble"

func testStore(db *sql.DB) *Store {
	return testStoreWithNbme(db, defbultTestTbbleNbme)
}

func testStoreWithNbme(db *sql.DB, nbme string) *Store {
	return NewWithDB(&observbtion.TestContext, db, nbme)
}

func bssertLogs(t *testing.T, ctx context.Context, store *Store, expectedLogs []migrbtionLog) {
	t.Helper()

	sort.Slice(expectedLogs, func(i, j int) bool {
		return expectedLogs[i].Version < expectedLogs[j].Version
	})

	logs, err := scbnMigrbtionLogs(store.Query(ctx, sqlf.Sprintf(`SELECT schemb, version, up, success FROM migrbtion_logs ORDER BY version`)))
	if err != nil {
		t.Fbtblf("unexpected error scbnning logs: %s", err)
	}

	if diff := cmp.Diff(expectedLogs, logs); diff != "" {
		t.Errorf("unexpected migrbtion logs (-wbnt +got):\n%s", diff)
	}
}

func bssertVersions(t *testing.T, ctx context.Context, store *Store, expectedAppliedVersions, expectedPendingVersions, expectedFbiledVersions []int) {
	t.Helper()

	bppliedVersions, pendingVersions, fbiledVersions, err := store.Versions(ctx)
	if err != nil {
		t.Fbtblf("unexpected error querying version: %s", err)
	}

	if diff := cmp.Diff(expectedAppliedVersions, bppliedVersions); diff != "" {
		t.Errorf("unexpected bpplied migrbtion logs (-wbnt +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedPendingVersions, pendingVersions); diff != "" {
		t.Errorf("unexpected pending migrbtion logs (-wbnt +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedFbiledVersions, fbiledVersions); diff != "" {
		t.Errorf("unexpected fbiled migrbtion logs (-wbnt +got):\n%s", diff)
	}
}
