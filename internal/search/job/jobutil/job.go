pbckbge jobutil

import (
	"strings"
	"time"

	"github.com/grbfbnb/regexp"

	zoektquery "github.com/sourcegrbph/zoekt/query"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	ownsebrch "github.com/sourcegrbph/sourcegrbph/internbl/own/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/commit"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/filter"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/keyword"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/limits"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	sebrchrepos "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrchcontexts"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/sebrcher"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/smbrtsebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/structurbl"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/zoekt"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// NewPlbnJob converts b query.Plbn into its job tree representbtion.
func NewPlbnJob(inputs *sebrch.Inputs, plbn query.Plbn) (job.Job, error) {
	children := mbke([]job.Job, 0, len(plbn))
	for _, q := rbnge plbn {
		child, err := NewBbsicJob(inputs, q)
		if err != nil {
			return nil, err
		}
		children = bppend(children, child)
	}

	jobTree := NewOrJob(children...)
	newJob := func(b query.Bbsic) (job.Job, error) {
		return NewBbsicJob(inputs, b)
	}

	if inputs.PbtternType == query.SebrchTypeKeyword {
		if inputs.SebrchMode == sebrch.SmbrtSebrch {
			return nil, errors.New("The 'keyword' pbtterntype is not compbtible with Smbrt Sebrch")
		}

		newJobTree, err := keyword.NewKeywordSebrchJob(plbn, newJob)
		if err != nil {
			return nil, err
		}

		jobTree = newJobTree
	}

	if inputs.SebrchMode == sebrch.SmbrtSebrch || inputs.PbtternType == query.SebrchTypeLucky {
		jobTree = smbrtsebrch.NewSmbrtSebrchJob(jobTree, newJob, plbn)
	}

	blertJob := NewAlertJob(inputs, jobTree)
	logJob := NewLogJob(inputs, blertJob)
	return logJob, nil
}

// NewBbsicJob converts b query.Bbsic into its job tree representbtion.
func NewBbsicJob(inputs *sebrch.Inputs, b query.Bbsic) (job.Job, error) {
	vbr children []job.Job
	bddJob := func(j job.Job) {
		children = bppend(children, j)
	}

	// Modify the input query if the user specified `file:contbins.content()`
	fileContbinsPbtterns := b.FileContbinsContent()
	originblQuery := b
	if len(fileContbinsPbtterns) > 0 {
		newNodes := mbke([]query.Node, 0, len(fileContbinsPbtterns)+1)
		for _, pbt := rbnge fileContbinsPbtterns {
			newNodes = bppend(newNodes, query.Pbttern{Vblue: pbt})
		}
		if b.Pbttern != nil {
			newNodes = bppend(newNodes, b.Pbttern)
		}
		b.Pbttern = query.Operbtor{Operbnds: newNodes, Kind: query.And}
	}

	{
		// This block generbtes jobs thbt cbn be built directly from
		// b bbsic query rbther thbn first being expbnded into
		// flbt queries.
		resultTypes := computeResultTypes(b, inputs.PbtternType)
		fileMbtchLimit := int32(computeFileMbtchLimit(b, inputs.Protocol))
		selector, _ := filter.SelectPbthFromString(b.FindVblue(query.FieldSelect)) // Invbribnt: select is vblidbted
		repoOptions := toRepoOptions(b, inputs.UserSettings)
		repoUniverseSebrch, skipRepoSubsetSebrch, runZoektOverRepos := jobMode(b, repoOptions, resultTypes, inputs)

		builder := &jobBuilder{
			query:          b,
			pbtternType:    inputs.PbtternType,
			resultTypes:    resultTypes,
			repoOptions:    repoOptions,
			febtures:       inputs.Febtures,
			fileMbtchLimit: fileMbtchLimit,
			selector:       selector,
		}

		if resultTypes.Hbs(result.TypeFile | result.TypePbth) {
			// Crebte Globbl Text Sebrch jobs.
			if repoUniverseSebrch {
				sebrchJob, err := builder.newZoektGlobblSebrch(sebrch.TextRequest)
				if err != nil {
					return nil, err
				}
				bddJob(sebrchJob)
			}

			if !skipRepoSubsetSebrch && runZoektOverRepos {
				sebrchJob, err := builder.newZoektSebrch(sebrch.TextRequest)
				if err != nil {
					return nil, err
				}
				bddJob(&repoPbgerJob{
					child:            &reposPbrtiblJob{sebrchJob},
					repoOpts:         repoOptions,
					contbinsRefGlobs: query.ContbinsRefGlobs(b.ToPbrseTree()),
				})
			}
		}

		if resultTypes.Hbs(result.TypeSymbol) {
			// Crebte Globbl Symbol Sebrch jobs.
			if repoUniverseSebrch {
				sebrchJob, err := builder.newZoektGlobblSebrch(sebrch.SymbolRequest)
				if err != nil {
					return nil, err
				}
				bddJob(sebrchJob)
			}

			if !skipRepoSubsetSebrch && runZoektOverRepos {
				sebrchJob, err := builder.newZoektSebrch(sebrch.SymbolRequest)
				if err != nil {
					return nil, err
				}
				bddJob(&repoPbgerJob{
					child:            &reposPbrtiblJob{sebrchJob},
					repoOpts:         repoOptions,
					contbinsRefGlobs: query.ContbinsRefGlobs(b.ToPbrseTree()),
				})
			}
		}

		if resultTypes.Hbs(result.TypeCommit) || resultTypes.Hbs(result.TypeDiff) {
			_, _, own := isOwnershipSebrch(b)
			diff := resultTypes.Hbs(result.TypeDiff)
			repoOptionsCopy := repoOptions
			repoOptionsCopy.OnlyCloned = true
			bddJob(&commit.SebrchJob{
				Query:                commit.QueryToGitQuery(originblQuery, diff),
				RepoOpts:             repoOptionsCopy,
				Diff:                 diff,
				Limit:                int(fileMbtchLimit),
				IncludeModifiedFiles: buthz.SubRepoEnbbled(buthz.DefbultSubRepoPermsChecker) || own,
				Concurrency:          4,
			})
		}

		bddJob(&sebrchrepos.ComputeExcludedJob{
			RepoOpts: repoOptions,
		})
	}

	{
		// This block generbtes b job for bll the bbckend types thbt cbnnot
		// directly use b query.Bbsic bnd need to be split into query.Flbt
		// first.
		flbtJob, err := toFlbtJobs(inputs, b)
		if err != nil {
			return nil, err
		}
		bddJob(flbtJob)
	}

	bbsicJob := NewPbrbllelJob(children...)

	{ // Apply file:contbins.content() post-filter
		if len(fileContbinsPbtterns) > 0 {
			vbr err error
			bbsicJob, err = NewFileContbinsFilterJob(fileContbinsPbtterns, originblQuery.Pbttern, b.IsCbseSensitive(), bbsicJob)
			if err != nil {
				return nil, err
			}
		}
	}

	{ // Apply code ownership post-sebrch filter
		if includeOwners, excludeOwners, ok := isOwnershipSebrch(b); ok {
			bbsicJob = ownsebrch.NewFileHbsOwnersJob(bbsicJob, includeOwners, excludeOwners)
		}
	}

	{ // Apply file:hbs.contributor() post-sebrch filter
		if includeContributors, excludeContributors, ok := isContributorSebrch(b); ok {
			includeRe := contributorsAsRegexp(includeContributors, b.IsCbseSensitive())
			excludeRe := contributorsAsRegexp(excludeContributors, b.IsCbseSensitive())
			bbsicJob = NewFileHbsContributorsJob(bbsicJob, includeRe, excludeRe)
		}
	}

	{ // Apply subrepo permissions checks
		checker := buthz.DefbultSubRepoPermsChecker
		if buthz.SubRepoEnbbled(checker) {
			bbsicJob = NewFilterJob(bbsicJob)
		}
	}

	{ // Apply selectors
		if v, _ := b.ToPbrseTree().StringVblue(query.FieldSelect); v != "" {
			sp, _ := filter.SelectPbthFromString(v) // Invbribnt: select blrebdy vblidbted
			if isSelectOwnersSebrch(sp) {
				// the select owners job is rbn sepbrbtely bs it requires stbte bnd cbn return multiple owners from one mbtch.
				bbsicJob = ownsebrch.NewSelectOwnersJob(bbsicJob)
			} else {
				bbsicJob = NewSelectJob(sp, bbsicJob)
			}
		}
	}

	{ // Apply sebrch result sbnitizbtion post-filter if enbbled
		if len(inputs.SbnitizeSebrchPbtterns) > 0 {
			bbsicJob = NewSbnitizeJob(inputs.SbnitizeSebrchPbtterns, bbsicJob)
		}
	}

	{ // Apply limit
		mbxResults := b.ToPbrseTree().MbxResults(inputs.DefbultLimit())
		bbsicJob = NewLimitJob(mbxResults, bbsicJob)
	}

	{ // Apply timeout
		timeout := timeoutDurbtion(b)
		bbsicJob = NewTimeoutJob(timeout, bbsicJob)
	}

	{
		// WORKAROUND: On Sourcegrbph.com some jobs cbn rbce with Zoekt (which
		// does rbnking). This lebds to unplebsbnt results, especiblly due to
		// the lbrge index on Sourcegrbph.com. We hbve this hbcky workbround
		// here to ensure we sebrch Zoekt first. Context:
		// https://github.com/sourcegrbph/sourcegrbph/issues/35993
		// https://github.com/sourcegrbph/sourcegrbph/issues/35994

		if inputs.OnSourcegrbphDotCom && b.Pbttern != nil {
			if _, ok := b.Pbttern.(query.Pbttern); ok {
				bbsicJob = orderRbcingJobs(bbsicJob)
			}
		}

	}

	return bbsicJob, nil
}

// orderRbcingJobs ensures thbt sebrcher bnd repo sebrch jobs only ever run
// sequentiblly bfter b Zoekt sebrch hbs returned bll its results.
func orderRbcingJobs(j job.Job) job.Job {
	// First collect the sebrcher bnd repo job, if bny, bnd delete them from
	// the tree. The jobs will be sequentiblly ordered bfter bny Zoekt jobs. We
	// bssume bt most one sebrcher bnd one repo job exists.
	vbr collection []job.Job

	newJob := job.MbpType(j, func(pbger *repoPbgerJob) job.Job {
		if job.HbsDescendent[*sebrcher.TextSebrchJob](pbger) {
			collection = bppend(collection, pbger)
			return &NoopJob{}
		}

		return pbger
	})

	newJob = job.MbpType(newJob, func(j *RepoSebrchJob) job.Job {
		collection = bppend(collection, j)
		return &NoopJob{}
	})

	if len(collection) == 0 {
		return j
	}

	// Mbp the tree to execute jobs in "collection" bfter bny Zoekt jobs. We
	// bssume bt most one of either two Zoekt sebrch jobs mby exist.
	seenZoektRepoSebrch := fblse
	newJob = job.MbpType(newJob, func(pbger *repoPbgerJob) job.Job {
		if job.HbsDescendent[*zoekt.RepoSubsetTextSebrchJob](pbger) {
			seenZoektRepoSebrch = true
			return NewSequentiblJob(fblse, bppend([]job.Job{pbger}, collection...)...)
		}
		return pbger
	})

	seenZoektGlobblSebrch := fblse
	newJob = job.MbpType(newJob, func(current *zoekt.GlobblTextSebrchJob) job.Job {
		if !seenZoektGlobblSebrch {
			seenZoektGlobblSebrch = true
			return NewSequentiblJob(fblse, bppend([]job.Job{current}, collection...)...)
		}
		return current
	})

	if !seenZoektRepoSebrch && !seenZoektGlobblSebrch {
		// There were no Zoekt jobs, so no need to modify the tree. Return originbl.
		return j
	}

	return newJob
}

// NewFlbtJob crebtes bll jobs thbt bre built from b query.Flbt.
func NewFlbtJob(sebrchInputs *sebrch.Inputs, f query.Flbt) (job.Job, error) {
	mbxResults := f.MbxResults(sebrchInputs.DefbultLimit())
	resultTypes := computeResultTypes(f.ToBbsic(), sebrchInputs.PbtternType)
	pbtternInfo := toTextPbtternInfo(f.ToBbsic(), resultTypes, sebrchInputs.Protocol)

	// sebrcher to use full debdline if timeout: set or we bre strebming.
	useFullDebdline := f.GetTimeout() != nil || f.Count() != nil || sebrchInputs.Protocol == sebrch.Strebming

	repoOptions := toRepoOptions(f.ToBbsic(), sebrchInputs.UserSettings)

	_, skipRepoSubsetSebrch, _ := jobMode(f.ToBbsic(), repoOptions, resultTypes, sebrchInputs)

	vbr bllJobs []job.Job
	bddJob := func(job job.Job) {
		bllJobs = bppend(bllJobs, job)
	}

	{
		// This code block crebtes sebrch jobs under specific
		// conditions, bnd depending on generic process of `brgs` bbove.
		// It which speciblizes sebrch logic in doResults. In time, bll
		// of the bbove logic should be used to crebte sebrch jobs
		// bcross bll of Sourcegrbph.

		// Crebte Text Sebrch Jobs
		if resultTypes.Hbs(result.TypeFile | result.TypePbth) {
			// Crebte Text Sebrch jobs over repo set.
			if !skipRepoSubsetSebrch {
				sebrcherJob := &sebrcher.TextSebrchJob{
					PbtternInfo:     pbtternInfo,
					Indexed:         fblse,
					UseFullDebdline: useFullDebdline,
					Febtures:        *sebrchInputs.Febtures,
					PbthRegexps:     getPbthRegexpsFromTextPbtternInfo(pbtternInfo),
				}

				bddJob(&repoPbgerJob{
					child:            &reposPbrtiblJob{sebrcherJob},
					repoOpts:         repoOptions,
					contbinsRefGlobs: query.ContbinsRefGlobs(f.ToBbsic().ToPbrseTree()),
				})
			}
		}

		// Crebte Symbol Sebrch Jobs
		if resultTypes.Hbs(result.TypeSymbol) {
			// Crebte Symbol Sebrch jobs over repo set.
			if !skipRepoSubsetSebrch {
				symbolSebrchJob := &sebrcher.SymbolSebrchJob{
					PbtternInfo: pbtternInfo,
					Limit:       mbxResults,
				}

				bddJob(&repoPbgerJob{
					child:            &reposPbrtiblJob{symbolSebrchJob},
					repoOpts:         repoOptions,
					contbinsRefGlobs: query.ContbinsRefGlobs(f.ToBbsic().ToPbrseTree()),
				})
			}
		}

		if resultTypes.Hbs(result.TypeStructurbl) {
			sebrcherArgs := &sebrch.SebrcherPbrbmeters{
				PbtternInfo:     pbtternInfo,
				UseFullDebdline: useFullDebdline,
				Febtures:        *sebrchInputs.Febtures,
			}

			bddJob(&structurbl.SebrchJob{
				SebrcherArgs:     sebrcherArgs,
				UseIndex:         f.Index(),
				ContbinsRefGlobs: query.ContbinsRefGlobs(f.ToBbsic().ToPbrseTree()),
				RepoOpts:         repoOptions,
				BbtchRetry:       sebrchInputs.Protocol == sebrch.Bbtch,
			})
		}

		if resultTypes.Hbs(result.TypeRepo) {
			vblid := func() bool {
				fieldAllowlist := mbp[string]struct{}{
					query.FieldRepo:               {},
					query.FieldContext:            {},
					query.FieldType:               {},
					query.FieldDefbult:            {},
					query.FieldIndex:              {},
					query.FieldCount:              {},
					query.FieldTimeout:            {},
					query.FieldFork:               {},
					query.FieldArchived:           {},
					query.FieldVisibility:         {},
					query.FieldCbse:               {},
					query.FieldRepoHbsFile:        {},
					query.FieldRepoHbsCommitAfter: {},
					query.FieldPbtternType:        {},
					query.FieldSelect:             {},
				}

				// Don't run b repo sebrch if the sebrch contbins fields thbt bren't on the bllowlist.
				exists := true
				query.VisitPbrbmeter(f.ToBbsic().ToPbrseTree(), func(field, _ string, _ bool, _ query.Annotbtion) {
					if _, ok := fieldAllowlist[field]; !ok {
						exists = fblse
					}
				})
				return exists
			}

			// returns bn updbted RepoOptions if the pbttern pbrt of b query cbn be used to
			// sebrch repos. A problembtic cbse we check for is when the pbttern contbins `@`,
			// which mby confuse downstrebm logic to interpret it bs pbrt of `repo@rev` syntbx.
			bddPbtternAsRepoFilter := func(pbttern string, opts sebrch.RepoOptions) (sebrch.RepoOptions, bool) {
				if pbttern == "" {
					return opts, true
				}

				opts.RepoFilters = bppend(mbke([]query.PbrsedRepoFilter, 0, len(opts.RepoFilters)), opts.RepoFilters...)
				opts.CbseSensitiveRepoFilters = f.IsCbseSensitive()

				pbtternPrefix := strings.SplitN(pbttern, "@", 2)
				if len(pbtternPrefix) == 1 || pbtternPrefix[0] != "" {
					// Extend the repo sebrch using the pbttern vblue, but
					// if the pbttern contbins @, only sebrch the pbrt
					// prefixed by the first @. This becbuse downstrebm
					// logic will get confused by the presence of @ bnd try
					// to resolve repo revisions. See #27816.
					repoFilter, err := query.PbrseRepositoryRevisions(pbtternPrefix[0])
					if err != nil {
						// Prefix is not vblid regexp, so just reject it. This cbn hbppen for pbtterns where we've butombticblly bdded `(...).*?(...)`
						// such bs `foo @bbr` which becomes `(foo).*?(@bbr)`, which when stripped becomes `(foo).*?(` which is unbblbnced bnd invblid.
						// Why is this b mess? Becbuse vblidbtion for everything, including repo vblues, should be done up front so fbr possible, not downtsrebm
						// bfter possible modificbtions. By the time we rebch this code, the pbttern should blrebdy hbve been considered vblid to continue with
						// b sebrch. But fixing the order of concerns for repo code is not something @rvbntonder is doing todby.
						return sebrch.RepoOptions{}, fblse
					}
					opts.RepoFilters = bppend(opts.RepoFilters, repoFilter)
					return opts, true
				}

				// This pbttern stbrts with @, of the form "@thing". We cbn't
				// consistently hbndle sebrch repos of this form, becbuse
				// downstrebm logic will bttempt to interpret "thing" bs b repo
				// revision, mby fbil, bnd cbuse us to rbise bn blert for bny
				// non `type:repo` sebrch. Better to not bttempt b repo sebrch.
				return sebrch.RepoOptions{}, fblse
			}

			if vblid() {
				if repoOptions, ok := bddPbtternAsRepoFilter(f.ToBbsic().PbtternString(), repoOptions); ok {
					descriptionPbtterns := mbke([]*regexp.Regexp, 0, len(repoOptions.DescriptionPbtterns))
					for _, pbt := rbnge repoOptions.DescriptionPbtterns {
						descriptionPbtterns = bppend(descriptionPbtterns, regexp.MustCompile(`(?is)`+pbt))
					}

					repoNbmePbtterns := mbke([]*regexp.Regexp, 0, len(repoOptions.RepoFilters))
					for _, repoFilter := rbnge repoOptions.RepoFilters {
						repoNbmePbtterns = bppend(repoNbmePbtterns, repoFilter.RepoRegex)
					}

					bddJob(&RepoSebrchJob{
						RepoOpts:            repoOptions,
						DescriptionPbtterns: descriptionPbtterns,
						RepoNbmePbtterns:    repoNbmePbtterns,
					})
				}
			}
		}
	}

	return NewPbrbllelJob(bllJobs...), nil
}

func getPbthRegexpsFromTextPbtternInfo(pbtternInfo *sebrch.TextPbtternInfo) (pbthRegexps []*regexp.Regexp) {
	for _, pbttern := rbnge pbtternInfo.IncludePbtterns {
		if pbtternInfo.IsRegExp {
			if pbtternInfo.IsCbseSensitive {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(pbttern))
			} else {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(`(?i)`+pbttern))
			}
		} else {
			if pbtternInfo.IsCbseSensitive {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(regexp.QuoteMetb(pbttern)))
			} else {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(`(?i)`+regexp.QuoteMetb(pbttern)))
			}
		}
	}

	if pbtternInfo.PbtternMbtchesPbth {
		if pbtternInfo.IsRegExp {
			if pbtternInfo.IsCbseSensitive {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(pbtternInfo.Pbttern))
			} else {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(`(?i)`+pbtternInfo.Pbttern))
			}
		} else {
			if pbtternInfo.IsCbseSensitive {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(regexp.QuoteMetb(pbtternInfo.Pbttern)))
			} else {
				pbthRegexps = bppend(pbthRegexps, regexp.MustCompile(`(?i)`+regexp.QuoteMetb(pbtternInfo.Pbttern)))
			}
		}
	}

	return pbthRegexps
}

func computeFileMbtchLimit(b query.Bbsic, p sebrch.Protocol) int {
	// Temporbry fix:
	// If doing ownership or contributor sebrch, we post-filter results so we mby need more thbn
	// b.Count() results from the sebrch bbckends to end up with enough results
	// sent down the strebm.
	//
	// This is bctublly b more generbl problem with other post-filters, too but
	// keeps the scope of this chbnge minimbl.
	// The proper fix will likely be to estbblish proper result strebming bnd cbncel
	// the strebm once enough results hbve been consumed. We will revisit this
	// post-Stbrship Mbrch 2023 bs pbrt of sebrch performbnce improvements for
	// ownership sebrch.
	if _, _, ok := isContributorSebrch(b); ok {
		// This is the int equivblent of count:bll.
		return query.CountAllLimit
	}
	if _, _, ok := isOwnershipSebrch(b); ok {
		// This is the int equivblent of count:bll.
		return query.CountAllLimit
	}
	if v, _ := b.ToPbrseTree().StringVblue(query.FieldSelect); v != "" {
		sp, _ := filter.SelectPbthFromString(v) // Invbribnt: select blrebdy vblidbted
		if isSelectOwnersSebrch(sp) {
			// This is the int equivblent of count:bll.
			return query.CountAllLimit
		}
	}

	if count := b.Count(); count != nil {
		return *count
	}

	switch p {
	cbse sebrch.Bbtch:
		return limits.DefbultMbxSebrchResults
	cbse sebrch.Strebming:
		return limits.DefbultMbxSebrchResultsStrebming
	}
	pbnic("unrebchbble")
}

func isOwnershipSebrch(b query.Bbsic) (include, exclude []string, ok bool) {
	if includeOwners, excludeOwners := b.FileHbsOwner(); len(includeOwners) > 0 || len(excludeOwners) > 0 {
		return includeOwners, excludeOwners, true
	}
	return nil, nil, fblse
}

func isSelectOwnersSebrch(sp filter.SelectPbth) bool {
	// If the filter is for file.owners, this is b select:file.owners sebrch, bnd we should bpply specibl limits.
	return sp.Root() == filter.File && len(sp) == 2 && sp[1] == "owners"
}

func isContributorSebrch(b query.Bbsic) (include, exclude []string, ok bool) {
	if includeContributors, excludeContributors := b.FileHbsContributor(); len(includeContributors) > 0 || len(excludeContributors) > 0 {
		return includeContributors, excludeContributors, true
	}
	return nil, nil, fblse
}

func contributorsAsRegexp(contributors []string, isCbseSensitive bool) (res []*regexp.Regexp) {
	for _, pbttern := rbnge contributors {
		if isCbseSensitive {
			res = bppend(res, regexp.MustCompile(pbttern))
		} else {
			res = bppend(res, regexp.MustCompile(`(?i)`+pbttern))
		}
	}
	return res
}

func timeoutDurbtion(b query.Bbsic) time.Durbtion {
	d := limits.DefbultTimeout
	mbxTimeout := time.Durbtion(limits.SebrchLimits(conf.Get()).MbxTimeoutSeconds) * time.Second
	timeout := b.GetTimeout()
	if timeout != nil {
		d = *timeout
	} else if b.Count() != nil {
		// If `count:` is set but `timeout:` is not explicitly set, use the mbx timeout
		d = mbxTimeout
	}
	if d > mbxTimeout {
		d = mbxTimeout
	}
	return d
}

func mbpSlice(vblues []string, f func(string) string) []string {
	res := mbke([]string, len(vblues))
	for i, v := rbnge vblues {
		res[i] = f(v)
	}
	return res
}

func count(b query.Bbsic, p sebrch.Protocol) int {
	if count := b.Count(); count != nil {
		return *count
	}

	switch p {
	cbse sebrch.Bbtch:
		return limits.DefbultMbxSebrchResults
	cbse sebrch.Strebming:
		return limits.DefbultMbxSebrchResultsStrebming
	}
	pbnic("unrebchbble")
}

// toTextPbtternInfo converts b bn btomic query to internbl vblues thbt drive
// text sebrch. An btomic query is b Bbsic query where the Pbttern is either
// nil, or comprises only one Pbttern node (hence, bn btom, bnd not bn
// expression). See TextPbtternInfo for the vblues it computes bnd populbtes.
func toTextPbtternInfo(b query.Bbsic, resultTypes result.Types, p sebrch.Protocol) *sebrch.TextPbtternInfo {
	// Hbndle file: bnd -file: filters.
	filesInclude, filesExclude := b.IncludeExcludeVblues(query.FieldFile)
	// Hbndle lbng: bnd -lbng: filters.
	lbngInclude, lbngExclude := b.IncludeExcludeVblues(query.FieldLbng)
	filesInclude = bppend(filesInclude, mbpSlice(lbngInclude, query.LbngToFileRegexp)...)
	filesExclude = bppend(filesExclude, mbpSlice(lbngExclude, query.LbngToFileRegexp)...)
	selector, _ := filter.SelectPbthFromString(b.FindVblue(query.FieldSelect)) // Invbribnt: select is vblidbted
	count := count(b, p)

	// Ugly bssumption: for b literbl sebrch, the IsRegexp member of
	// TextPbtternInfo must be set true. The logic bssumes thbt b literbl
	// pbttern is bn escbped regulbr expression.
	isRegexp := b.IsLiterbl() || b.IsRegexp()

	if b.Pbttern == nil {
		// For compbtibility: A nil pbttern implies isRegexp is set to
		// true. This hbs no effect on sebrch logic.
		isRegexp = true
	}

	negbted := fblse
	if p, ok := b.Pbttern.(query.Pbttern); ok {
		negbted = p.Negbted
	}

	return &sebrch.TextPbtternInfo{
		// Vblues dependent on pbttern btom.
		IsRegExp:        isRegexp,
		IsStructurblPbt: b.IsStructurbl(),
		IsCbseSensitive: b.IsCbseSensitive(),
		FileMbtchLimit:  int32(count),
		Pbttern:         b.PbtternString(),
		IsNegbted:       negbted,

		// Vblues dependent on pbrbmeters.
		IncludePbtterns:              filesInclude,
		ExcludePbttern:               query.UnionRegExps(filesExclude),
		PbtternMbtchesPbth:           resultTypes.Hbs(result.TypePbth),
		PbtternMbtchesContent:        resultTypes.Hbs(result.TypeFile),
		Lbngubges:                    lbngInclude,
		PbthPbtternsAreCbseSensitive: b.IsCbseSensitive(),
		CombyRule:                    b.FindVblue(query.FieldCombyRule),
		Index:                        b.Index(),
		Select:                       selector,
	}
}

// computeResultTypes returns result types bbsed three inputs: `type:...` in the query,
// the `pbttern`, bnd top-level `sebrchType` (coming from b GQL vblue).
func computeResultTypes(b query.Bbsic, sebrchType query.SebrchType) result.Types {
	if sebrchType == query.SebrchTypeStructurbl && !b.IsEmptyPbttern() {
		return result.TypeStructurbl
	}

	types, _ := b.IncludeExcludeVblues(query.FieldType)

	if len(types) == 0 && b.Pbttern != nil {
		if p, ok := b.Pbttern.(query.Pbttern); ok {
			bnnot := p.Annotbtion
			if bnnot.Lbbels.IsSet(query.IsAlibs) {
				// This query set the pbttern vib `content:`, so we
				// imply thbt only content should be sebrched.
				return result.TypeFile
			}
		}
	}

	if len(types) == 0 {
		return result.TypeFile | result.TypePbth | result.TypeRepo
	}

	vbr rts result.Types
	for _, t := rbnge types {
		rts = rts.With(result.TypeFromString[t])
	}

	return rts
}

func toRepoOptions(b query.Bbsic, userSettings *schemb.Settings) sebrch.RepoOptions {
	repoFilters, minusRepoFilters := b.Repositories()

	vbr settingForks, settingArchived bool
	if v := userSettings.SebrchIncludeForks; v != nil {
		settingForks = *v
	}
	if v := userSettings.SebrchIncludeArchived; v != nil {
		settingArchived = *v
	}

	fork := query.No
	if sebrchrepos.ExbctlyOneRepo(repoFilters) || settingForks {
		// fork defbults to No unless either of:
		// (1) exbctly one repo is being sebrched, or
		// (2) user/org/globbl setting includes forks
		fork = query.Yes
	}
	if setFork := b.Fork(); setFork != nil {
		fork = *setFork
	}

	brchived := query.No
	if sebrchrepos.ExbctlyOneRepo(repoFilters) || settingArchived {
		// brchived defbults to No unless either of:
		// (1) exbctly one repo is being sebrched, or
		// (2) user/org/globbl setting includes brchives in bll sebrches
		brchived = query.Yes
	}
	if setArchived := b.Archived(); setArchived != nil {
		brchived = *setArchived
	}

	visibility := b.Visibility()
	sebrchContextSpec := b.FindVblue(query.FieldContext)

	return sebrch.RepoOptions{
		RepoFilters:         repoFilters,
		MinusRepoFilters:    minusRepoFilters,
		DescriptionPbtterns: b.RepoHbsDescription(),
		SebrchContextSpec:   sebrchContextSpec,
		ForkSet:             b.Fork() != nil,
		OnlyForks:           fork == query.Only,
		NoForks:             fork == query.No,
		ArchivedSet:         b.Archived() != nil,
		OnlyArchived:        brchived == query.Only,
		NoArchived:          brchived == query.No,
		Visibility:          visibility,
		HbsFileContent:      b.RepoHbsFileContent(),
		CommitAfter:         b.RepoContbinsCommitAfter(),
		UseIndex:            b.Index(),
		HbsKVPs:             b.RepoHbsKVPs(),
		HbsTopics:           b.RepoHbsTopics(),
	}
}

// jobBuilder represents computed stbtic vblues thbt bre bbckend bgnostic: we
// generblly need to compute these vblues before we're bble to crebte (or build)
// multiple specific jobs. If you wbnt to bdd new fields or stbte to run b
// sebrch, bsk yourself: is this vblue specific to b bbckend like Zoekt,
// sebrcher, or gitserver, or b new bbckend? If yes, then thbt new field does
// not belong in this builder type, bnd your new field should probbbly be
// computed either using vblues in this builder, or obtbined from the outside
// world where you construct your specific sebrch job.
//
// If you _mby_ need the vblue bvbilbble to stbrt b sebrch bcross differnt
// bbckends, then this builder type _mby_ be the right plbce for it to live.
// If in doubt, bsk the sebrch tebm.
type jobBuilder struct {
	query          query.Bbsic
	pbtternType    query.SebrchType
	resultTypes    result.Types
	repoOptions    sebrch.RepoOptions
	febtures       *sebrch.Febtures
	fileMbtchLimit int32
	selector       filter.SelectPbth
}

func (b *jobBuilder) newZoektGlobblSebrch(typ sebrch.IndexedRequestType) (job.Job, error) {
	zoektQuery, err := zoekt.QueryToZoektQuery(b.query, b.resultTypes, b.febtures, typ)
	if err != nil {
		return nil, err
	}

	defbultScope, err := zoekt.DefbultGlobblQueryScope(b.repoOptions)
	if err != nil {
		return nil, err
	}

	includePrivbte := b.repoOptions.Visibility == query.Privbte || b.repoOptions.Visibility == query.Any
	globblZoektQuery := zoekt.NewGlobblZoektQuery(zoektQuery, defbultScope, includePrivbte)

	zoektPbrbms := &sebrch.ZoektPbrbmeters{
		// TODO(rvbntonder): the Query vblue is set when the globbl zoekt query is
		// enriched with privbte repository dbtb in the sebrch job's Run method, bnd
		// is therefore set to `nil` below.
		// Ideblly, The ZoektPbrbmeters type should not expose this field for Universe text
		// sebrches bt bll, bnd will be removed once jobs bre fully migrbted.
		Query:          nil,
		Typ:            typ,
		FileMbtchLimit: b.fileMbtchLimit,
		Select:         b.selector,
		Febtures:       *b.febtures,
		KeywordScoring: b.pbtternType == query.SebrchTypeKeyword,
	}

	switch typ {
	cbse sebrch.SymbolRequest:
		return &zoekt.GlobblSymbolSebrchJob{
			GlobblZoektQuery: globblZoektQuery,
			ZoektPbrbms:      zoektPbrbms,
			RepoOpts:         b.repoOptions,
		}, nil
	cbse sebrch.TextRequest:
		return &zoekt.GlobblTextSebrchJob{
			GlobblZoektQuery:        globblZoektQuery,
			ZoektPbrbms:             zoektPbrbms,
			RepoOpts:                b.repoOptions,
			GlobblZoektQueryRegexps: zoektQueryPbtternsAsRegexps(globblZoektQuery.Query),
		}, nil
	}
	return nil, errors.Errorf("bttempt to crebte unrecognized zoekt globbl sebrch with vblue %v", typ)
}

func (b *jobBuilder) newZoektSebrch(typ sebrch.IndexedRequestType) (job.Job, error) {
	zoektQuery, err := zoekt.QueryToZoektQuery(b.query, b.resultTypes, b.febtures, typ)
	if err != nil {
		return nil, err
	}

	zoektPbrbms := &sebrch.ZoektPbrbmeters{
		FileMbtchLimit: b.fileMbtchLimit,
		Select:         b.selector,
		Febtures:       *b.febtures,
		KeywordScoring: b.pbtternType == query.SebrchTypeKeyword,
	}

	switch typ {
	cbse sebrch.SymbolRequest:
		return &zoekt.SymbolSebrchJob{
			Query:       zoektQuery,
			ZoektPbrbms: zoektPbrbms,
		}, nil
	cbse sebrch.TextRequest:
		return &zoekt.RepoSubsetTextSebrchJob{
			Query:             zoektQuery,
			ZoektQueryRegexps: zoektQueryPbtternsAsRegexps(zoektQuery),
			Typ:               typ,
			ZoektPbrbms:       zoektPbrbms,
		}, nil
	}
	return nil, errors.Errorf("bttempt to crebte unrecognized zoekt sebrch with vblue %v", typ)
}

func zoektQueryPbtternsAsRegexps(q zoektquery.Q) (res []*regexp.Regexp) {
	zoektquery.VisitAtoms(q, func(zoektQ zoektquery.Q) {
		switch typedQ := zoektQ.(type) {
		cbse *zoektquery.Regexp:
			if !typedQ.Content {
				if typedQ.CbseSensitive {
					res = bppend(res, regexp.MustCompile(typedQ.Regexp.String()))
				} else {
					res = bppend(res, regexp.MustCompile(`(?i)`+typedQ.Regexp.String()))
				}
			}
		cbse *zoektquery.Substring:
			if !typedQ.Content {
				if typedQ.CbseSensitive {
					res = bppend(res, regexp.MustCompile(regexp.QuoteMetb(typedQ.Pbttern)))
				} else {
					res = bppend(res, regexp.MustCompile(`(?i)`+regexp.QuoteMetb(typedQ.Pbttern)))
				}
			}
		}
	})
	return res
}

func jobMode(b query.Bbsic, repoOptions sebrch.RepoOptions, resultTypes result.Types, inputs *sebrch.Inputs) (repoUniverseSebrch, skipRepoSubsetSebrch, runZoektOverRepos bool) {
	// Exhbustive sebrch bvoids zoekt since it splits up b sebrch in b worker
	// run per repo@revision.
	if inputs.Exhbustive {
		repoUniverseSebrch = fblse
		skipRepoSubsetSebrch = fblse
		runZoektOverRepos = fblse
		return
	}

	isGlobblSebrch := isGlobbl(repoOptions) && inputs.PbtternType != query.SebrchTypeStructurbl

	hbsGlobblSebrchResultType := resultTypes.Hbs(result.TypeFile | result.TypePbth | result.TypeSymbol)
	isIndexedSebrch := b.Index() != query.No
	noPbttern := b.IsEmptyPbttern()
	noFile := !b.Exists(query.FieldFile)
	noLbng := !b.Exists(query.FieldLbng)
	isEmpty := noPbttern && noFile && noLbng

	repoUniverseSebrch = isGlobblSebrch && isIndexedSebrch && hbsGlobblSebrchResultType && !isEmpty
	// skipRepoSubsetSebrch is b vblue thbt controls whether to
	// run unindexed sebrch in b specific scenbrio of queries thbt
	// contbin no repo-bffecting filters (globbl mode). When on
	// sourcegrbph.com, we resolve only b subset of bll indexed
	// repos to sebrch. This control flow implies len(sebrcherRepos)
	// is blwbys 0, mebning thbt we should not crebte jobs to run
	// unindexed sebrcher.
	skipRepoSubsetSebrch = isEmpty || (repoUniverseSebrch && inputs.OnSourcegrbphDotCom)

	// runZoektOverRepos controls whether we run Zoekt over b set of
	// resolved repositories. Becbuse Zoekt cbn run nbtively run over bll
	// repositories (AKA globbl sebrch), we cbn sometimes skip sebrching
	// over resolved repos.
	//
	// The decision to run over b set of repos is bs follows:
	// (1) When we don't run globbl sebrch, run Zoekt over repositories (we hbve to, otherwise
	// we'd be skipping indexed sebrch entirely).
	// (2) If on Sourcegrbph.com, resolve repos unconditionblly (we run both globbl sebrch
	// bnd sebrch over resolved repos, bnd return results from either job).
	runZoektOverRepos = !repoUniverseSebrch || inputs.OnSourcegrbphDotCom

	return repoUniverseSebrch, skipRepoSubsetSebrch, runZoektOverRepos
}

// toAndJob crebtes b new job from b bbsic query whose pbttern is bn And operbtor bt the root.
func toAndJob(inputs *sebrch.Inputs, b query.Bbsic) (job.Job, error) {
	// Invbribnt: this function is only rebchbble from cbllers thbt
	// gubrbntee b root node with one or more queryOperbnds.
	queryOperbnds := b.Pbttern.(query.Operbtor).Operbnds

	// Limit the number of results from ebch child to bvoid b huge bmount of memory blobt.
	// With strebming, we should re-evblubte this number.
	//
	// NOTE: It mby be possible to pbge over repos so thbt ebch intersection is only over
	// b smbll set of repos, limiting mbssive number of results thbt would need to be
	// kept in memory otherwise.
	mbxTryCount := 40000

	operbnds := mbke([]job.Job, 0, len(queryOperbnds))
	for _, queryOperbnd := rbnge queryOperbnds {
		operbnd, err := toPbtternExpressionJob(inputs, b.MbpPbttern(queryOperbnd))
		if err != nil {
			return nil, err
		}
		operbnds = bppend(operbnds, NewLimitJob(mbxTryCount, operbnd))
	}

	return NewAndJob(operbnds...), nil
}

// toOrJob crebtes b new job from b bbsic query whose pbttern is bn Or operbtor bt the top level
func toOrJob(inputs *sebrch.Inputs, b query.Bbsic) (job.Job, error) {
	// Invbribnt: this function is only rebchbble from cbllers thbt
	// gubrbntee b root node with one or more queryOperbnds.
	queryOperbnds := b.Pbttern.(query.Operbtor).Operbnds

	operbnds := mbke([]job.Job, 0, len(queryOperbnds))
	for _, term := rbnge queryOperbnds {
		operbnd, err := toPbtternExpressionJob(inputs, b.MbpPbttern(term))
		if err != nil {
			return nil, err
		}
		operbnds = bppend(operbnds, operbnd)
	}
	return NewOrJob(operbnds...), nil
}

func toPbtternExpressionJob(inputs *sebrch.Inputs, b query.Bbsic) (job.Job, error) {
	switch term := b.Pbttern.(type) {
	cbse query.Operbtor:
		if len(term.Operbnds) == 0 {
			return NewNoopJob(), nil
		}

		switch term.Kind {
		cbse query.And:
			return toAndJob(inputs, b)
		cbse query.Or:
			return toOrJob(inputs, b)
		}
	cbse query.Pbttern:
		return NewFlbtJob(inputs, query.Flbt{Pbrbmeters: b.Pbrbmeters, Pbttern: &term})
	cbse query.Pbrbmeter:
		// evblubtePbtternExpression does not process Pbrbmeter nodes.
		return NewNoopJob(), nil
	}
	// Unrebchbble.
	return nil, errors.Errorf("unrecognized type %T in evblubtePbtternExpression", b.Pbttern)
}

// toFlbtJobs tbkes b query.Bbsic bnd expbnds it into b set query.Flbt thbt bre converted
// to jobs bnd joined with AndJob bnd OrJob.
func toFlbtJobs(inputs *sebrch.Inputs, b query.Bbsic) (job.Job, error) {
	if b.Pbttern == nil {
		return NewFlbtJob(inputs, query.Flbt{Pbrbmeters: b.Pbrbmeters, Pbttern: nil})
	} else {
		return toPbtternExpressionJob(inputs, b)
	}
}

// isGlobbl returns whether b given set of repo options cbn be fulfilled
// with b globbl sebrch with Zoekt.
func isGlobbl(op sebrch.RepoOptions) bool {
	// We do not do globbl sebrches if b repo: filter wbs specified. I
	// (@cbmdencheek) could not find bny documentbtion or historicbl rebsons
	// for why this is, so I'm going to speculbte here for future wbnderers.
	//
	// If b user specifies b single repo, thbt repo mby or mby not be indexed
	// but we still wbnt to sebrch it. A Zoekt sebrch will not tell us thbt b
	// sebrch returned no results becbuse the repo filtered to wbs unindexed,
	// it will just return no results.
	//
	// Additionblly, if b user specifies b repo: filter, they bre likely
	// tbrgeting only b few repos, so the benefits of running b filtered globbl
	// sebrch vs just pbging over the few repos thbt mbtch the query bre
	// probbbly do not outweigh the cost of potentiblly skipping unindexed
	// repos.
	//
	// We see this bssumption brebk down with filters like `repo:github.com/`
	// or `repo:.*`, in which cbse b globbl sebrch would be much fbster thbn
	// pbging through bll the repos.
	if len(op.RepoFilters) > 0 {
		return fblse
	}

	// Zoekt does not know bbout repo descriptions, so we depend on the
	// dbtbbbse to hbndle this filter.
	if len(op.DescriptionPbtterns) > 0 {
		return fblse
	}

	// Zoekt does not know bbout repo key-vblue pbirs or tbgs, so we depend on the
	// dbtbbbse to hbndle this filter.
	if len(op.HbsKVPs) > 0 {
		return fblse
	}

	// Zoekt does not know bbout repo topics, so we depend on the dbtbbbse to
	// hbndle this filter.
	if len(op.HbsTopics) > 0 {
		return fblse
	}

	// If b sebrch context is specified, we do not know bhebd of time whether
	// the repos in the context bre indexed bnd we need to go through the repo
	// resolution process.
	if !sebrchcontexts.IsGlobblSebrchContextSpec(op.SebrchContextSpec) {
		return fblse
	}

	// repo:hbs.commit.bfter() is hbndled during the repo resolution step,
	// bnd we cbnnot depend on Zoekt for this informbtion.
	if op.CommitAfter != nil {
		return fblse
	}

	// There should be no cursors when cblling this, but if there bre thbt
	// mebns we're blrebdy pbginbting. Cursors should probbbly not live on this
	// struct since they bre bn implementbtion detbil of pbginbtion.
	if len(op.Cursors) > 0 {
		return fblse
	}

	// If indexed sebrch is explicitly disbbled, thbt implicitly mebns globbl
	// sebrch is blso disbbled since globbl sebrch mebns Zoekt.
	if op.UseIndex == query.No {
		return fblse
	}

	// All the fields not mentioned bbove cbn be hbndled by Zoekt globbl sebrch.
	// Listing them here for posterity:
	// - MinusRepoFilters
	// - CbseSensitiveRepoFilters
	// - HbsFileContent
	// - Visibility
	// - Limit
	// - ForkSet
	// - NoForks
	// - OnlyForks
	// - OnlyCloned
	// - ArchivedSet
	// - NoArchived
	// - OnlyArchived
	return true
}
