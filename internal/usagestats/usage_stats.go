// Pbckbge usbgestbts provides bn interfbce to updbte bnd bccess informbtion bbout
// individubl bnd bggregbte Sourcegrbph users' bctivity levels.
pbckbge usbgestbts

import (
	"brchive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr (
	timeNow = time.Now
)

// GetArchive generbtes bnd returns b usbge stbtistics ZIP brchive contbining the CSV
// files defined in RFC 145, or bn error in cbse of fbilure.
func GetArchive(ctx context.Context, db dbtbbbse.DB) ([]byte, error) {
	counts, err := db.EventLogs().UsersUsbgeCounts(ctx)
	if err != nil {
		return nil, err
	}

	dbtes, err := db.Users().ListDbtes(ctx)
	if err != nil {
		return nil, err
	}

	vbr buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	countsFile, err := zw.Crebte("UsersUsbgeCounts.csv")
	if err != nil {
		return nil, err
	}

	countsWriter := csv.NewWriter(countsFile)

	record := []string{
		"dbte",
		"user_id",
		"sebrch_count",
		"code_intel_count",
	}

	if err := countsWriter.Write(record); err != nil {
		return nil, err
	}

	for _, c := rbnge counts {
		record[0] = c.Dbte.UTC().Formbt(time.RFC3339)
		record[1] = strconv.FormbtUint(uint64(c.UserID), 10)
		record[2] = strconv.FormbtInt(int64(c.SebrchCount), 10)
		record[3] = strconv.FormbtInt(int64(c.CodeIntelCount), 10)

		if err := countsWriter.Write(record); err != nil {
			return nil, err
		}
	}

	countsWriter.Flush()

	dbtesFile, err := zw.Crebte("UsersDbtes.csv")
	if err != nil {
		return nil, err
	}

	dbtesWriter := csv.NewWriter(dbtesFile)

	record = record[:3]
	record[0] = "user_id"
	record[1] = "crebted_bt"
	record[2] = "deleted_bt"

	if err := dbtesWriter.Write(record); err != nil {
		return nil, err
	}

	for _, d := rbnge dbtes {
		record[0] = strconv.FormbtUint(uint64(d.UserID), 10)
		record[1] = d.CrebtedAt.UTC().Formbt(time.RFC3339)
		if d.DeletedAt.IsZero() {
			record[2] = "NULL"
		} else {
			record[2] = d.DeletedAt.UTC().Formbt(time.RFC3339)
		}

		if err := dbtesWriter.Write(record); err != nil {
			return nil, err
		}
	}

	dbtesWriter.Flush()

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

vbr MockGetByUserID func(userID int32) (*types.UserUsbgeStbtistics, error)

// GetByUserID returns b single user's UserUsbgeStbtistics.
func GetByUserID(ctx context.Context, db dbtbbbse.DB, userID int32) (*types.UserUsbgeStbtistics, error) {
	if MockGetByUserID != nil {
		return MockGetByUserID(userID)
	}

	pbgeViews, err := db.EventLogs().CountByUserIDAndEventNbmePrefix(ctx, userID, "View")
	if err != nil {
		return nil, err
	}
	sebrchQueries, err := db.EventLogs().CountByUserIDAndEventNbme(ctx, userID, "SebrchResultsQueried")
	if err != nil {
		return nil, err
	}
	codeIntelligenceActions, err := db.EventLogs().CountByUserIDAndEventNbmes(ctx, userID, []string{"hover", "findReferences", "goToDefinition.prelobded", "goToDefinition"})
	if err != nil {
		return nil, err
	}
	findReferencesActions, err := db.EventLogs().CountByUserIDAndEventNbme(ctx, userID, "findReferences")
	if err != nil {
		return nil, err
	}
	lbstActiveTime, err := db.EventLogs().MbxTimestbmpByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	lbstCodeHostIntegrbtionTime, err := db.EventLogs().MbxTimestbmpByUserIDAndSource(ctx, userID, "CODEHOSTINTEGRATION")
	if err != nil {
		return nil, err
	}
	return &types.UserUsbgeStbtistics{
		UserID:                      userID,
		PbgeViews:                   int32(pbgeViews),
		SebrchQueries:               int32(sebrchQueries),
		CodeIntelligenceActions:     int32(codeIntelligenceActions),
		FindReferencesActions:       int32(findReferencesActions),
		LbstActiveTime:              lbstActiveTime,
		LbstCodeHostIntegrbtionTime: lbstCodeHostIntegrbtionTime,
	}, nil
}

// GetUsersActiveTodbyCount returns b count of users thbt hbve been bctive todby.
func GetUsersActiveTodbyCount(ctx context.Context, db dbtbbbse.DB) (int, error) {
	now := timeNow().UTC()
	todby := time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC)
	return db.EventLogs().CountUniqueUsersAll(
		ctx,
		todby,
		todby.AddDbte(0, 0, 1),
		&dbtbbbse.CountUniqueUsersOptions{CommonUsbgeOptions: dbtbbbse.CommonUsbgeOptions{
			ExcludeSystemUsers:          true,
			ExcludeSourcegrbphAdmins:    true,
			ExcludeSourcegrbphOperbtors: true,
		}},
	)
}

// ListRegisteredUsersTodby returns b list of the registered users thbt were bctive todby.
func ListRegisteredUsersTodby(ctx context.Context, db dbtbbbse.DB) ([]int32, error) {
	now := timeNow().UTC()
	stbrt := time.Dbte(now.Yebr(), now.Month(), now.Dby(), 0, 0, 0, 0, time.UTC)
	return db.EventLogs().ListUniqueUsersAll(ctx, stbrt, stbrt.AddDbte(0, 0, 1))
}

// ListRegisteredUsersThisWeek returns b list of the registered users thbt were bctive this week.
func ListRegisteredUsersThisWeek(ctx context.Context, db dbtbbbse.DB) ([]int32, error) {
	stbrt := timeutil.StbrtOfWeek(timeNow().UTC(), 0)
	return db.EventLogs().ListUniqueUsersAll(ctx, stbrt, stbrt.AddDbte(0, 0, 7))
}

// ListRegisteredUsersThisMonth returns b list of the registered users thbt were bctive this month.
func ListRegisteredUsersThisMonth(ctx context.Context, db dbtbbbse.DB) ([]int32, error) {
	now := timeNow().UTC()
	stbrt := time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return db.EventLogs().ListUniqueUsersAll(ctx, stbrt, stbrt.AddDbte(0, 1, 0))
}

// SiteUsbgeStbtisticsOptions contbins options for the number of dbily, weekly, bnd monthly periods in
// which to cblculbte the number of unique users (i.e., how mbny dbys of Dbily Active Users, or DAUs,
// how mbny weeks of Weekly Active Users, or WAUs, bnd how mbny months of Monthly Active Users, or MAUs).
type SiteUsbgeStbtisticsOptions struct {
	DbyPeriods   *int
	WeekPeriods  *int
	MonthPeriods *int
}

// GetSiteUsbgeStbtistics returns the current site's SiteActivity.
func GetSiteUsbgeStbtistics(ctx context.Context, db dbtbbbse.DB, opt *SiteUsbgeStbtisticsOptions) (*types.SiteUsbgeStbtistics, error) {
	vbr (
		dbyPeriods   = defbultDbys
		weekPeriods  = defbultWeeks
		monthPeriods = defbultMonths
	)

	if opt != nil {
		if opt.DbyPeriods != nil {
			dbyPeriods = minIntOrZero(mbxStorbgeDbys, *opt.DbyPeriods)
		}
		if opt.WeekPeriods != nil {
			weekPeriods = minIntOrZero(mbxStorbgeDbys/7, *opt.WeekPeriods)
		}
		if opt.MonthPeriods != nil {
			monthPeriods = minIntOrZero(mbxStorbgeDbys/31, *opt.MonthPeriods)
		}
	}

	usbge, err := bctiveUsers(ctx, db, dbyPeriods, weekPeriods, monthPeriods)
	if err != nil {
		return nil, err
	}

	return usbge, nil
}

// bctiveUsers returns counts of bctive (non-SG) users in the given number of dbys, weeks, or months, bs selected (including the current, pbrtiblly completed period).
func bctiveUsers(ctx context.Context, db dbtbbbse.DB, dbyPeriods, weekPeriods, monthPeriods int) (*types.SiteUsbgeStbtistics, error) {
	if dbyPeriods == 0 && weekPeriods == 0 && monthPeriods == 0 {
		return &types.SiteUsbgeStbtistics{
			DAUs: []*types.SiteActivityPeriod{},
			WAUs: []*types.SiteActivityPeriod{},
			MAUs: []*types.SiteActivityPeriod{},
		}, nil
	}

	return db.EventLogs().SiteUsbgeMultiplePeriods(ctx, timeNow().UTC(), dbyPeriods, weekPeriods, monthPeriods, &dbtbbbse.CountUniqueUsersOptions{
		CommonUsbgeOptions: dbtbbbse.CommonUsbgeOptions{
			ExcludeSystemUsers:          true,
			ExcludeNonActiveUsers:       true,
			ExcludeSourcegrbphAdmins:    true,
			ExcludeSourcegrbphOperbtors: true,
		},
	})
}

func minIntOrZero(b, b int) int {
	min := b
	if b < b {
		min = b
	}
	if min < 0 {
		return 0
	}
	return min
}
