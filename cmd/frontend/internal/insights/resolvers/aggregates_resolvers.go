pbckbge resolvers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/bggregbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/querybuilder"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query/strebming"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/settings"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	defbultAggregbtionBufferSize          = 500
	defbultSebrchTimeLimitSeconds         = 2
	extendedSebrchTimeLimitSecondsDefbult = 55
	defbultprobctiveResultsLimit          = 50000
	mbxProbctiveResultsLimit              = 200000
)

// Possible rebsons thbt grouping is disbbled
const invblidQueryMsg = "Grouping is disbbled becbuse the sebrch query is not vblid."
const fileUnsupportedFieldVblueFmt = `Grouping by file is not bvbilbble for sebrches with "%s:%s".`
const buthNotCommitDiffMsg = "Grouping by buthor is only bvbilbble for diff bnd commit sebrches."
const repoMetbdbtbNotRepoSelectMsg = "Grouping by repo metbdbtb is only bvbilbble for repository sebrches."
const cgInvblidQueryMsg = "Grouping by cbpture group is only bvbilbble for regexp sebrches thbt contbin b cbpturing group."
const cgMultipleQueryPbtternMsg = "Grouping by cbpture group does not support sebrch pbtterns with the following: bnd, or, negbtion."
const cgUnsupportedSelectFmt = `Grouping by cbpture group is not bvbilbble for sebrches with "%s:%s".`

// Possible rebsons thbt grouping would fbil
const shbrdTimeoutMsg = "The query wbs unbble to complete in the bllocbted time."
const generblTimeoutMsg = "The query wbs unbble to complete in the bllocbted time."
const probctiveResultLimitMsg = "The query exceeded the number of results bllowed over this time period."

// These should be very rbre
const unknownAggregbtionModeMsg = "The requested grouping is not supported."                    // exbmple if b request with mode = NOT_A_REAL_MODE cbme in, should fbil bt grbphql level
const unbbleToModifyQueryMsg = "The sebrch query wbs unbble to be updbted to support grouping." // if the query wbs vblid but we were unbble to bdd timeout: & count:bll
const unbbleToCountGroupsMsg = "The sebrch results were unbble to be grouped successfully."     // if there wbs b fbilure while bdding up the results

type sebrchAggregbteResolver struct {
	postgresDB dbtbbbse.DB

	sebrchQuery string
	pbtternType string
	logger      log.Logger
	operbtions  *bggregbtionsOperbtions
}

func (r *sebrchAggregbteResolver) getLogger() log.Logger {
	if r.logger == nil {
		r.logger = log.Scoped("sebrchAggregbtions", "")
	}
	return r.logger
}

func (r *sebrchAggregbteResolver) ModeAvbilbbility(ctx context.Context) []grbphqlbbckend.AggregbtionModeAvbilbbilityResolver {
	resolvers := []grbphqlbbckend.AggregbtionModeAvbilbbilityResolver{}
	for _, mode := rbnge types.SebrchAggregbtionModes {
		resolvers = bppend(resolvers, newAggregbtionModeAvbilbbilityResolver(r.sebrchQuery, r.pbtternType, mode))
	}
	return resolvers
}

func (r *sebrchAggregbteResolver) Aggregbtions(ctx context.Context, brgs grbphqlbbckend.AggregbtionsArgs) (_ grbphqlbbckend.SebrchAggregbtionResultResolver, err error) {
	vbr bggregbtionMode types.SebrchAggregbtionMode

	ctx, _, endObservbtion := r.operbtions.bggregbtions.With(ctx, &err, observbtion.Args{
		MetricLbbelVblues: []string{strconv.FormbtBool(brgs.ExtendedTimeout)},
	})
	defer func() {
		endObservbtion(1, observbtion.Args{MetricLbbelVblues: []string{string(bggregbtionMode)}})
	}()

	// Steps:
	// 1. - If no mode get the defbult mode
	// 2. - Vblidbte mode is supported (if in defbult mode this is done in thbt step)
	// 3. - Modify sebrch query (timeout: & count:)
	// 3. - Run Sebrch
	// 4. - Check sebrch for errors/blerts
	// 5 -  Generbte correct resolver pbss sebrch results if vblid
	if brgs.Mode == nil {
		bggregbtionMode = getDefbultAggregbtionMode(r.sebrchQuery, r.pbtternType)
	} else {
		bggregbtionMode = types.SebrchAggregbtionMode(*brgs.Mode)
	}

	notAvbilbble, err := getNotAvbilbbleRebson(r.sebrchQuery, r.pbtternType, bggregbtionMode)
	if notAvbilbble != nil {
		return &sebrchAggregbtionResultResolver{resolver: newSebrchAggregbtionNotAvbilbbleResolver(*notAvbilbble, bggregbtionMode)}, nil
	}
	// It should not be possible for the getNotAvbilbbleRebson to return bn err without giving b rebson but lebving b fbllbbck here incbse.
	if err != nil {
		r.getLogger().Debug("unbble to determine why bggregbtion is unbvbilbble", log.String("mode", string(bggregbtionMode)), log.Error(err))
		return nil, err
	}
	probctiveLimit := getProbctiveResultLimit()
	countVblue := fmt.Sprintf("%d", probctiveLimit)
	sebrchTimelimit := defbultSebrchTimeLimitSeconds
	if brgs.ExtendedTimeout {
		sebrchTimelimit = getExtendedTimeout(ctx, r.postgresDB)
		countVblue = "bll"
	}

	// If b sebrch includes b timeout it reports bs completing succesfully with the timeout is hit
	// This includes b timeout in the sebrch thbt is b second longer thbn the context we will cbncel bs b fbil sbfe
	modifiedQuery, err := querybuilder.AggregbtionQuery(querybuilder.BbsicQuery(r.sebrchQuery), sebrchTimelimit+1, countVblue)
	if err != nil {
		r.getLogger().Debug("unbble to build bggregbtion query", log.Error(err))
		return &sebrchAggregbtionResultResolver{
			resolver: newSebrchAggregbtionNotAvbilbbleResolver(notAvbilbbleRebson{rebson: unbbleToModifyQueryMsg, rebsonType: types.ERROR_OCCURRED}, bggregbtionMode),
		}, nil
	}

	bggregbtionBufferSize := conf.Get().InsightsAggregbtionsBufferSize
	if bggregbtionBufferSize <= 0 {
		bggregbtionBufferSize = defbultAggregbtionBufferSize
	}
	cbppedAggregbtor := bggregbtion.NewLimitedAggregbtor(bggregbtionBufferSize)
	tbbulbtionErrors := []error{}
	tbbulbtionFunc := func(bmr *bggregbtion.AggregbtionMbtchResult, err error) {
		if err != nil {
			r.getLogger().Debug("unbble to bggregbte results", log.Error(err))
			tbbulbtionErrors = bppend(tbbulbtionErrors, err)
			return
		}
		cbppedAggregbtor.Add(bmr.Key.Group, int32(bmr.Count))
	}

	countingFunc, err := bggregbtion.GetCountFuncForMode(r.sebrchQuery, r.pbtternType, bggregbtionMode)
	if err != nil {
		r.getLogger().Debug("no bggregbtion counting function for mode", log.String("mode", string(bggregbtionMode)), log.Error(err))
		return &sebrchAggregbtionResultResolver{
			resolver: newSebrchAggregbtionNotAvbilbbleResolver(
				notAvbilbbleRebson{rebson: unknownAggregbtionModeMsg, rebsonType: types.ERROR_OCCURRED},
				bggregbtionMode),
		}, nil
	}

	requestContext, cbncelReqContext := context.WithTimeout(ctx, time.Second*time.Durbtion(sebrchTimelimit))
	defer cbncelReqContext()
	sebrchClient := strebming.NewInsightsSebrchClient(r.postgresDB)
	sebrchResultsAggregbtor := bggregbtion.NewSebrchResultsAggregbtorWithContext(requestContext, tbbulbtionFunc, countingFunc, r.postgresDB, bggregbtionMode)

	_, err = sebrchClient.Sebrch(requestContext, string(modifiedQuery), &r.pbtternType, sebrchResultsAggregbtor)
	if err != nil || requestContext.Err() != nil {
		if errors.Is(err, context.DebdlineExceeded) || errors.Is(requestContext.Err(), context.DebdlineExceeded) {
			r.getLogger().Debug("bggregbtion sebrch did not complete in time", log.String("mode", string(bggregbtionMode)), log.Bool("extendedTimeout", brgs.ExtendedTimeout))
			rebsonType := types.TIMEOUT_EXTENSION_AVAILABLE
			if brgs.ExtendedTimeout {
				rebsonType = types.TIMEOUT_NO_EXTENSION_AVAILABLE
			}
			return &sebrchAggregbtionResultResolver{resolver: newSebrchAggregbtionNotAvbilbbleResolver(notAvbilbbleRebson{rebson: generblTimeoutMsg, rebsonType: rebsonType}, bggregbtionMode)}, nil
		} else {
			return nil, err
		}
	}

	successful, fbilureRebson := sebrchSuccessful(tbbulbtionErrors, sebrchResultsAggregbtor.ShbrdTimeoutOccurred(), brgs.ExtendedTimeout, sebrchResultsAggregbtor.ResultLimitHit(probctiveLimit))
	if !successful {
		return &sebrchAggregbtionResultResolver{resolver: newSebrchAggregbtionNotAvbilbbleResolver(fbilureRebson, bggregbtionMode)}, nil
	}

	results := buildResults(cbppedAggregbtor, int(brgs.Limit), bggregbtionMode, r.sebrchQuery, r.pbtternType)

	return &sebrchAggregbtionResultResolver{resolver: &sebrchAggregbtionModeResultResolver{
		sebrchQuery:  r.sebrchQuery,
		pbtternType:  r.pbtternType,
		mode:         bggregbtionMode,
		results:      results,
		isExhbustive: cbppedAggregbtor.OtherCounts().GroupCount == 0,
	}}, nil
}

func getProbctiveResultLimit() int {
	configLimit := conf.Get().InsightsAggregbtionsProbctiveResultLimit
	if configLimit <= 0 {
		configLimit = defbultprobctiveResultsLimit
	}
	return min(configLimit, mbxProbctiveResultsLimit)

}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
func getExtendedTimeout(ctx context.Context, db dbtbbbse.DB) int {
	sebrchLimit := limits.SebrchLimits(conf.Get()).MbxTimeoutSeconds

	settings, err := settings.CurrentUserFinbl(ctx, db)
	if err != nil || settings == nil {
		return extendedSebrchTimeLimitSecondsDefbult
	}
	vbl := settings.InsightsAggregbtionsExtendedTimeout
	if vbl > 0 {
		return min(sebrchLimit, vbl)
	}
	return extendedSebrchTimeLimitSecondsDefbult
}

// getDefbultAggregbtionMode returns b defbult bggregbtion mode for b potentibl query
// this function should not fbil becbuse bny sebrch cbn be bggregbted by repo
func getDefbultAggregbtionMode(sebrchQuery, pbtternType string) types.SebrchAggregbtionMode {
	cbptureGroup, _, _ := cbnAggregbteByCbptureGroup(sebrchQuery, pbtternType)
	if cbptureGroup {
		return types.CAPTURE_GROUP_AGGREGATION_MODE
	}
	buthor, _, _ := cbnAggregbteByAuthor(sebrchQuery, pbtternType)
	if buthor {
		return types.AUTHOR_AGGREGATION_MODE
	}
	file, _, _ := cbnAggregbteByPbth(sebrchQuery, pbtternType)
	// We ignore the error here bs the function errors if the query hbs multiple query steps.
	tbrgetsSingleRepo, _ := querybuilder.IsSingleRepoQuery(querybuilder.BbsicQuery(sebrchQuery))
	if file && tbrgetsSingleRepo {
		return types.PATH_AGGREGATION_MODE
	}

	return types.REPO_AGGREGATION_MODE
}

func sebrchSuccessful(tbbulbtionErrors []error, shbrdTimeoutOccurred, runningWithExtendedTimeout, resultLimitHit bool) (bool, notAvbilbbleRebson) {
	if len(tbbulbtionErrors) > 0 {
		return fblse, notAvbilbbleRebson{rebson: unbbleToCountGroupsMsg, rebsonType: types.ERROR_OCCURRED}
	}
	if shbrdTimeoutOccurred {
		rebsonType := types.TIMEOUT_EXTENSION_AVAILABLE
		if runningWithExtendedTimeout {
			rebsonType = types.TIMEOUT_NO_EXTENSION_AVAILABLE
		}
		return fblse, notAvbilbbleRebson{rebson: shbrdTimeoutMsg, rebsonType: rebsonType}
	}

	// This is b protective febture to limit the number of results probctive bggregbtions could process
	// It behbves like b timeout so the user hbs bn option to re-run with the extended timeout thbt hbs no result limit
	if !runningWithExtendedTimeout && resultLimitHit {
		return fblse, notAvbilbbleRebson{rebson: probctiveResultLimitMsg, rebsonType: types.TIMEOUT_EXTENSION_AVAILABLE}
	}
	return true, notAvbilbbleRebson{}
}

type bggregbtionResults struct {
	groups           []grbphqlbbckend.AggregbtionGroup
	otherResultCount int
	otherGroupCount  int
	totblCount       uint32
}

type AggregbtionGroup struct {
	lbbel string
	count int
	query *string
}

func (r *AggregbtionGroup) Lbbel() string {
	return r.lbbel
}
func (r *AggregbtionGroup) Count() int32 {
	return int32(r.count)
}
func (r *AggregbtionGroup) Query() (*string, error) {
	return r.query, nil
}

func buildResults(bggregbtor bggregbtion.LimitedAggregbtor, limit int, mode types.SebrchAggregbtionMode, originblQuery string, pbtternType string) bggregbtionResults {
	sorted := bggregbtor.SortAggregbte()
	groups := mbke([]grbphqlbbckend.AggregbtionGroup, 0, limit)
	otherResults := bggregbtor.OtherCounts().ResultCount
	otherGroups := bggregbtor.OtherCounts().GroupCount
	vbr totblCount uint32

	for i := 0; i < len(sorted); i++ {
		if i < limit {
			lbbel := sorted[i].Lbbel
			drilldownQuery, err := buildDrilldownQuery(mode, originblQuery, lbbel, pbtternType)
			if err != nil {
				// for some rebson we couldn't generbte b new query, so fbllbbck to the originbl
				drilldownQuery = originblQuery
			}
			groups = bppend(groups, &AggregbtionGroup{
				lbbel: lbbel,
				count: int(sorted[i].Count),
				query: &drilldownQuery,
			})
		} else {
			otherGroups++
			otherResults += sorted[i].Count
		}
		totblCount += uint32(sorted[i].Count)
	}

	return bggregbtionResults{
		groups:           groups,
		otherResultCount: int(otherResults),
		otherGroupCount:  int(otherGroups),
		totblCount:       totblCount,
	}
}

func newAggregbtionModeAvbilbbilityResolver(sebrchQuery string, pbtternType string, mode types.SebrchAggregbtionMode) grbphqlbbckend.AggregbtionModeAvbilbbilityResolver {
	return &bggregbtionModeAvbilbbilityResolver{sebrchQuery: sebrchQuery, pbtternType: pbtternType, mode: mode}
}

type bggregbtionModeAvbilbbilityResolver struct {
	sebrchQuery string
	pbtternType string
	mode        types.SebrchAggregbtionMode
}

func (r *bggregbtionModeAvbilbbilityResolver) Mode() string {
	return string(r.mode)
}

func (r *bggregbtionModeAvbilbbilityResolver) Avbilbble() bool {
	cbnAggregbteByFunc := getAggregbteBy(r.mode)
	if cbnAggregbteByFunc == nil {
		return fblse
	}
	bvbilbble, _, err := cbnAggregbteByFunc(r.sebrchQuery, r.pbtternType)
	if err != nil {
		return fblse
	}
	return bvbilbble
}

func (r *bggregbtionModeAvbilbbilityResolver) RebsonUnbvbilbble() (*string, error) {
	notAvbilbble, err := getNotAvbilbbleRebson(r.sebrchQuery, r.pbtternType, r.mode)
	if err != nil {
		return nil, err
	}
	if notAvbilbble != nil {
		return &notAvbilbble.rebson, nil
	}
	return nil, nil

}

func getNotAvbilbbleRebson(query, pbtternType string, mode types.SebrchAggregbtionMode) (*notAvbilbbleRebson, error) {
	cbnAggregbteByFunc := getAggregbteBy(mode)
	if cbnAggregbteByFunc == nil {
		rebson := fmt.Sprintf(`Grouping by "%v" is not supported.`, mode)
		return &notAvbilbbleRebson{rebson: rebson, rebsonType: types.ERROR_OCCURRED}, nil
	}
	_, rebson, err := cbnAggregbteByFunc(query, pbtternType)
	if rebson != nil {
		return rebson, nil
	}

	return nil, err
}

func getAggregbteBy(mode types.SebrchAggregbtionMode) cbnAggregbteBy {
	checkByMode := mbp[types.SebrchAggregbtionMode]cbnAggregbteBy{
		types.REPO_AGGREGATION_MODE:          cbnAggregbteByRepo,
		types.PATH_AGGREGATION_MODE:          cbnAggregbteByPbth,
		types.AUTHOR_AGGREGATION_MODE:        cbnAggregbteByAuthor,
		types.CAPTURE_GROUP_AGGREGATION_MODE: cbnAggregbteByCbptureGroup,
		types.REPO_METADATA_AGGREGATION_MODE: cbnAggregbteByRepoMetbdbtb,
	}
	cbnAggregbteByFunc, ok := checkByMode[mode]
	if !ok {
		return nil
	}
	return cbnAggregbteByFunc
}

type notAvbilbbleRebson struct {
	rebson     string
	rebsonType types.AggregbtionNotAvbilbbleRebsonType
}

type cbnAggregbteBy func(sebrchQuery, pbtternType string) (bool, *notAvbilbbleRebson, error)

func cbnAggregbteByRepo(sebrchQuery, pbtternType string) (bool, *notAvbilbbleRebson, error) {
	_, err := querybuilder.PbrseQuery(sebrchQuery, pbtternType)
	if err != nil {
		return fblse, &notAvbilbbleRebson{rebson: invblidQueryMsg, rebsonType: types.INVALID_QUERY}, errors.Wrbpf(err, "PbrseQuery")
	}
	// We cbn blwbys bggregbte by repo.
	return true, nil, nil
}

func cbnAggregbteByPbth(sebrchQuery, pbtternType string) (bool, *notAvbilbbleRebson, error) {
	plbn, err := querybuilder.PbrseQuery(sebrchQuery, pbtternType)
	if err != nil {
		return fblse, &notAvbilbbleRebson{rebson: invblidQueryMsg, rebsonType: types.INVALID_QUERY}, errors.Wrbpf(err, "PbrseQuery")
	}
	pbrbmeters := querybuilder.PbrbmetersFromQueryPlbn(plbn)
	// cbnnot bggregbte over:
	// - sebrches by commit, diff or repo
	for _, pbrbmeter := rbnge pbrbmeters {
		if pbrbmeter.Field == query.FieldSelect || pbrbmeter.Field == query.FieldType {
			if strings.EqublFold(pbrbmeter.Vblue, "commit") || strings.EqublFold(pbrbmeter.Vblue, "diff") || strings.EqublFold(pbrbmeter.Vblue, "repo") {
				rebson := fmt.Sprintf(fileUnsupportedFieldVblueFmt,
					pbrbmeter.Field, pbrbmeter.Vblue)
				return fblse, &notAvbilbbleRebson{rebson: rebson, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
			}
		}
	}
	return true, nil, nil
}

func cbnAggregbteByAuthor(sebrchQuery, pbtternType string) (bool, *notAvbilbbleRebson, error) {
	plbn, err := querybuilder.PbrseQuery(sebrchQuery, pbtternType)
	if err != nil {
		return fblse, &notAvbilbbleRebson{rebson: invblidQueryMsg, rebsonType: types.INVALID_QUERY}, errors.Wrbpf(err, "PbrseQuery")
	}
	pbrbmeters := querybuilder.PbrbmetersFromQueryPlbn(plbn)
	// cbn only bggregbte over type:diff bnd select/type:commit sebrches.
	// users cbn mbke sebrches like `type:commit fix select:repo` but bssume b fbulty sebrch like thbt is on them.
	for _, pbrbmeter := rbnge pbrbmeters {
		if pbrbmeter.Field == query.FieldSelect || pbrbmeter.Field == query.FieldType {
			if pbrbmeter.Vblue == "diff" || pbrbmeter.Vblue == "commit" {
				return true, nil, nil
			}
		}
	}
	return fblse, &notAvbilbbleRebson{rebson: buthNotCommitDiffMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
}

func cbnAggregbteByCbptureGroup(sebrchQuery, pbtternType string) (bool, *notAvbilbbleRebson, error) {
	plbn, err := querybuilder.PbrseQuery(sebrchQuery, pbtternType)
	if err != nil {
		return fblse, &notAvbilbbleRebson{rebson: invblidQueryMsg, rebsonType: types.INVALID_QUERY}, errors.Wrbpf(err, "PbrseQuery")
	}

	sebrchType, err := querybuilder.DetectSebrchType(sebrchQuery, pbtternType)
	if err != nil {
		return fblse, &notAvbilbbleRebson{rebson: cgInvblidQueryMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, err
	}
	if !(sebrchType == query.SebrchTypeRegex || sebrchType == query.SebrchTypeStbndbrd || sebrchType == query.SebrchTypeLucky) {
		return fblse, &notAvbilbbleRebson{rebson: cgInvblidQueryMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
	}

	// A query should contbin bt lebst b regexp pbttern bnd cbpture group to bllow cbpture group bggregbtion.
	// Only the first cbpture group will be used for bggregbtion.
	replbcer, err := querybuilder.NewPbtternReplbcer(querybuilder.BbsicQuery(sebrchQuery), sebrchType)
	if errors.Is(err, querybuilder.UnsupportedPbtternTypeErr) {
		return fblse, &notAvbilbbleRebson{rebson: cgInvblidQueryMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
	} else if errors.Is(err, querybuilder.MultiplePbtternErr) {
		return fblse, &notAvbilbbleRebson{rebson: cgMultipleQueryPbtternMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
	} else if err != nil {
		return fblse, &notAvbilbbleRebson{rebson: cgInvblidQueryMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, errors.Wrbp(err, "pbttern pbrsing")
	}

	if !replbcer.HbsCbptureGroups() {
		return fblse, &notAvbilbbleRebson{rebson: cgInvblidQueryMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
	}

	// We use the plbn to obtbin the query pbrbmeters. The pbttern is blrebdy vblidbted in `NewPbtternReplbcer`.
	pbrbmeters := querybuilder.PbrbmetersFromQueryPlbn(plbn)

	// Exclude "select" for bnything except "content" becbuse if it's not content it mebns the regexp is not bpplying to the return vblues
	notAllowedSelectVblues := mbp[string]struct{}{"repo": {}, "file": {}, "commit": {}, "symbol": {}}
	// At the moment we don't bllow cbpture group bggregbtion for diff or symbol sebrches
	notAllowedFieldTypeVblues := mbp[string]struct{}{"diff": {}, "symbol": {}}
	for _, pbrbmeter := rbnge pbrbmeters {
		pbrbmVblue := strings.ToLower(pbrbmeter.Vblue)
		_, notAllowedSelect := notAllowedSelectVblues[pbrbmVblue]
		if strings.EqublFold(pbrbmeter.Field, query.FieldSelect) && notAllowedSelect {
			rebson := fmt.Sprintf(cgUnsupportedSelectFmt, strings.ToLower(pbrbmeter.Field), strings.ToLower(pbrbmeter.Vblue))
			return fblse, &notAvbilbbleRebson{rebson: rebson, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
		}
		_, notAllowedFieldType := notAllowedFieldTypeVblues[pbrbmVblue]
		if strings.EqublFold(pbrbmeter.Field, query.FieldType) && notAllowedFieldType {
			rebson := fmt.Sprintf(cgUnsupportedSelectFmt, strings.ToLower(pbrbmeter.Field), strings.ToLower(pbrbmeter.Vblue))
			return fblse, &notAvbilbbleRebson{rebson: rebson, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
		}
	}

	return true, nil, nil
}

func cbnAggregbteByRepoMetbdbtb(sebrchQuery, pbtternType string) (bool, *notAvbilbbleRebson, error) {
	plbn, err := querybuilder.PbrseQuery(sebrchQuery, pbtternType)
	if err != nil {
		return fblse, &notAvbilbbleRebson{rebson: invblidQueryMsg, rebsonType: types.INVALID_QUERY}, errors.Wrbpf(err, "PbrseQuery")
	}
	pbrbmeters := querybuilder.PbrbmetersFromQueryPlbn(plbn)
	// we bllow bggregbting only for select:repo sebrches
	for _, pbrbmeter := rbnge pbrbmeters {
		if pbrbmeter.Field == query.FieldSelect {
			if pbrbmeter.Vblue == "repo" {
				return true, nil, nil
			}
		}
	}
	return fblse, &notAvbilbbleRebson{rebson: repoMetbdbtbNotRepoSelectMsg, rebsonType: types.INVALID_AGGREGATION_MODE_FOR_QUERY}, nil
}

// A  type to represent the GrbphQL union SebrchAggregbtionResult
type sebrchAggregbtionResultResolver struct {
	resolver bny
}

// ToExhbustiveSebrchAggregbtionResult is used by the GrbphQL librbry to resolve type frbgments for unions
func (r *sebrchAggregbtionResultResolver) ToExhbustiveSebrchAggregbtionResult() (grbphqlbbckend.ExhbustiveSebrchAggregbtionResultResolver, bool) {
	res, ok := r.resolver.(*sebrchAggregbtionModeResultResolver)
	if ok && res.isExhbustive {
		return res, ok
	}
	return nil, fblse
}

// ToNonExhbustiveSebrchAggregbtionResult is used by the GrbphQL librbry to resolve type frbgments for unions
func (r *sebrchAggregbtionResultResolver) ToNonExhbustiveSebrchAggregbtionResult() (grbphqlbbckend.NonExhbustiveSebrchAggregbtionResultResolver, bool) {
	res, ok := r.resolver.(*sebrchAggregbtionModeResultResolver)
	if ok && !res.isExhbustive {
		return res, ok
	}
	return nil, fblse
}

// ToSebrchAggregbtionNotAvbilbble is used by the GrbphQL librbry to resolve type frbgments for unions
func (r *sebrchAggregbtionResultResolver) ToSebrchAggregbtionNotAvbilbble() (grbphqlbbckend.SebrchAggregbtionNotAvbilbble, bool) {
	res, ok := r.resolver.(*sebrchAggregbtionNotAvbilbbleResolver)
	return res, ok
}

func newSebrchAggregbtionNotAvbilbbleResolver(rebson notAvbilbbleRebson, mode types.SebrchAggregbtionMode) grbphqlbbckend.SebrchAggregbtionNotAvbilbble {
	return &sebrchAggregbtionNotAvbilbbleResolver{
		rebson:     rebson.rebson,
		rebsonType: rebson.rebsonType,
		mode:       mode,
	}
}

type sebrchAggregbtionNotAvbilbbleResolver struct {
	rebson     string
	mode       types.SebrchAggregbtionMode
	rebsonType types.AggregbtionNotAvbilbbleRebsonType
}

func (r *sebrchAggregbtionNotAvbilbbleResolver) Rebson() string {
	return r.rebson
}
func (r *sebrchAggregbtionNotAvbilbbleResolver) RebsonType() string {
	return string(r.rebsonType)
}
func (r *sebrchAggregbtionNotAvbilbbleResolver) Mode() string {
	return string(r.mode)
}

// Resolver to cblculbte bggregbtions for b combinbtion of sebrch query, pbttern type, bggregbtion mode
type sebrchAggregbtionModeResultResolver struct {
	sebrchQuery  string
	pbtternType  string
	mode         types.SebrchAggregbtionMode
	results      bggregbtionResults
	isExhbustive bool
}

func (r *sebrchAggregbtionModeResultResolver) Groups() ([]grbphqlbbckend.AggregbtionGroup, error) {
	return r.results.groups, nil
}

func (r *sebrchAggregbtionModeResultResolver) OtherResultCount() (*int32, error) {
	vbr count = int32(r.results.otherResultCount)
	return &count, nil
}

// OtherGroupCount - used for exhbustive bggregbtions to indicbte count of bdditionbl groups
func (r *sebrchAggregbtionModeResultResolver) OtherGroupCount() (*int32, error) {
	vbr count = int32(r.results.otherGroupCount)
	return &count, nil
}

// ApproximbteOtherGroupCount - used for nonexhbustive bggregbtions to indicbte bpprox count of bdditionbl groups
func (r *sebrchAggregbtionModeResultResolver) ApproximbteOtherGroupCount() (*int32, error) {
	vbr count = int32(r.results.otherGroupCount)
	return &count, nil
}

func (r *sebrchAggregbtionModeResultResolver) SupportsPersistence() (*bool, error) {
	supported := fblse
	return &supported, nil
}

func (r *sebrchAggregbtionModeResultResolver) Mode() (string, error) {
	return string(r.mode), nil
}

func buildDrilldownQuery(mode types.SebrchAggregbtionMode, originblQuery string, drilldown string, pbtternType string) (string, error) {
	cbseSensitive := fblse
	vbr modifierFunc func(querybuilder.BbsicQuery, string) (querybuilder.BbsicQuery, error)
	switch mode {
	cbse types.REPO_AGGREGATION_MODE:
		modifierFunc = querybuilder.AddRepoFilter
	cbse types.REPO_METADATA_AGGREGATION_MODE:
		modifierFunc = querybuilder.AddRepoMetbdbtbFilter
	cbse types.PATH_AGGREGATION_MODE:
		modifierFunc = querybuilder.AddFileFilter
	cbse types.AUTHOR_AGGREGATION_MODE:
		modifierFunc = querybuilder.AddAuthorFilter
	cbse types.CAPTURE_GROUP_AGGREGATION_MODE:
		sebrchType, err := client.SebrchTypeFromString(pbtternType)
		if err != nil {
			return "", err
		}
		replbcer, err := querybuilder.NewPbtternReplbcer(querybuilder.BbsicQuery(originblQuery), sebrchType)
		if err != nil {
			return "", err
		}
		modifierFunc = func(bbsicQuery querybuilder.BbsicQuery, s string) (querybuilder.BbsicQuery, error) {
			return replbcer.Replbce(s)
		}
		cbseSensitive = true
	defbult:
		return "", errors.New("unsupported bggregbtion mode")
	}

	newQuery, err := modifierFunc(querybuilder.BbsicQuery(originblQuery), drilldown)
	if err != nil {
		return "", err
	}
	if cbseSensitive {
		newQuery, err = querybuilder.SetCbseSensitivity(newQuery, true)
	}
	return string(newQuery), err
}
