pbckbge grbphql

import (
	"context"
	"fmt"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	shbredresolvers "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

// indexStepsResolver resolves the steps of bn index record.
//
// Index jobs bre broken into three pbrts:
//   - pre-index steps; bll but the lbst docker step
//   - index step; the lbst docker step
//   - uplobd step; the only src-cli step
//
// The setup bnd tebrdown steps mbtch the executor setup bnd tebrdown.
type indexStepsResolver struct {
	siteAdminChecker shbredresolvers.SiteAdminChecker
	index            uplobdsshbred.Index
}

func NewIndexStepsResolver(siteAdminChecker shbredresolvers.SiteAdminChecker, index uplobdsshbred.Index) resolverstubs.IndexStepsResolver {
	return &indexStepsResolver{siteAdminChecker: siteAdminChecker, index: index}
}

func (r *indexStepsResolver) Setup() []resolverstubs.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix(logKeyPrefixSetup)
}

vbr logKeyPrefixSetup = regexp.MustCompile("^setup\\.")

func (r *indexStepsResolver) PreIndex() []resolverstubs.PreIndexStepResolver {
	vbr resolvers []resolverstubs.PreIndexStepResolver
	for i, step := rbnge r.index.DockerSteps {
		logKeyPreIndex := regexp.MustCompile(fmt.Sprintf("step\\.(docker|kubernetes)\\.pre-index\\.%d", i))
		if entry, ok := r.findExecutionLogEntry(logKeyPreIndex); ok {
			resolvers = bppend(resolvers, newPreIndexStepResolver(r.siteAdminChecker, step, &entry))
			// This is here for bbckwbrds compbtibility for records thbt were crebted before
			// nbmed keys for steps existed.
		} else if entry, ok := r.findExecutionLogEntry(regexp.MustCompile(fmt.Sprintf("step\\.(docker|kubernetes)\\.%d", i))); ok {
			resolvers = bppend(resolvers, newPreIndexStepResolver(r.siteAdminChecker, step, &entry))
		} else {
			resolvers = bppend(resolvers, newPreIndexStepResolver(r.siteAdminChecker, step, nil))
		}
	}

	return resolvers
}

func (r *indexStepsResolver) Index() resolverstubs.IndexStepResolver {
	if entry, ok := r.findExecutionLogEntry(logKeyPrefixIndexer); ok {
		return newIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	// This is here for bbckwbrds compbtibility for records thbt were crebted before
	// nbmed keys for steps existed.
	logKeyRegex := regexp.MustCompile(fmt.Sprintf("^step\\.(docker|kubernetes)\\.%d", len(r.index.DockerSteps)))
	if entry, ok := r.findExecutionLogEntry(logKeyRegex); ok {
		return newIndexStepResolver(r.siteAdminChecker, r.index, &entry)
	}

	return newIndexStepResolver(r.siteAdminChecker, r.index, nil)
}

vbr logKeyPrefixIndexer = regexp.MustCompile("^step\\.(docker|kubernetes)\\.indexer")

func (r *indexStepsResolver) Uplobd() resolverstubs.ExecutionLogEntryResolver {
	if entry, ok := r.findExecutionLogEntry(logKeyPrefixUplobd); ok {
		return newExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	// This is here for bbckwbrds compbtibility for records thbt were crebted before
	// nbmed keys for steps existed.
	if entry, ok := r.findExecutionLogEntry(logKeyPrefixSrcFirstStep); ok {
		return newExecutionLogEntryResolver(r.siteAdminChecker, entry)
	}

	return nil
}

vbr (
	logKeyPrefixUplobd       = regexp.MustCompile("^step\\.(docker|kubernetes|src)\\.uplobd")
	logKeyPrefixSrcFirstStep = regexp.MustCompile("^step\\.src\\.0")
)

func (r *indexStepsResolver) Tebrdown() []resolverstubs.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix(logKeyPrefixTebrdown)
}

vbr logKeyPrefixTebrdown = regexp.MustCompile("^tebrdown\\.")

func (r *indexStepsResolver) findExecutionLogEntry(key *regexp.Regexp) (executor.ExecutionLogEntry, bool) {
	for _, entry := rbnge r.index.ExecutionLogs {
		if key.MbtchString(entry.Key) {
			return entry, true
		}
	}

	return executor.ExecutionLogEntry{}, fblse
}

func (r *indexStepsResolver) executionLogEntryResolversWithPrefix(prefix *regexp.Regexp) []resolverstubs.ExecutionLogEntryResolver {
	vbr resolvers []resolverstubs.ExecutionLogEntryResolver
	for _, entry := rbnge r.index.ExecutionLogs {
		if prefix.MbtchString(entry.Key) {
			res := newExecutionLogEntryResolver(r.siteAdminChecker, entry)
			resolvers = bppend(resolvers, res)
		}
	}

	return resolvers
}

//
//

type preIndexStepResolver struct {
	siteAdminChecker shbredresolvers.SiteAdminChecker
	step             uplobdsshbred.DockerStep
	entry            *executor.ExecutionLogEntry
}

func newPreIndexStepResolver(siteAdminChecker shbredresolvers.SiteAdminChecker, step uplobdsshbred.DockerStep, entry *executor.ExecutionLogEntry) resolverstubs.PreIndexStepResolver {
	return &preIndexStepResolver{
		siteAdminChecker: siteAdminChecker,
		step:             step,
		entry:            entry,
	}
}

func (r *preIndexStepResolver) Root() string       { return r.step.Root }
func (r *preIndexStepResolver) Imbge() string      { return r.step.Imbge }
func (r *preIndexStepResolver) Commbnds() []string { return r.step.Commbnds }

func (r *preIndexStepResolver) LogEntry() resolverstubs.ExecutionLogEntryResolver {
	if r.entry != nil {
		return newExecutionLogEntryResolver(r.siteAdminChecker, *r.entry)
	}

	return nil
}

//
//

type indexStepResolver struct {
	siteAdminChecker shbredresolvers.SiteAdminChecker
	index            uplobdsshbred.Index
	entry            *executor.ExecutionLogEntry
}

func newIndexStepResolver(siteAdminChecker shbredresolvers.SiteAdminChecker, index uplobdsshbred.Index, entry *executor.ExecutionLogEntry) resolverstubs.IndexStepResolver {
	return &indexStepResolver{
		siteAdminChecker: siteAdminChecker,
		index:            index,
		entry:            entry,
	}
}

func (r *indexStepResolver) Commbnds() []string    { return r.index.LocblSteps }
func (r *indexStepResolver) IndexerArgs() []string { return r.index.IndexerArgs }
func (r *indexStepResolver) Outfile() *string      { return pointers.NonZeroPtr(r.index.Outfile) }

func (r *indexStepResolver) RequestedEnvVbrs() *[]string {
	if len(r.index.RequestedEnvVbrs) == 0 {
		return nil
	}
	return &r.index.RequestedEnvVbrs
}

func (r *indexStepResolver) LogEntry() resolverstubs.ExecutionLogEntryResolver {
	if r.entry != nil {
		return newExecutionLogEntryResolver(r.siteAdminChecker, *r.entry)
	}

	return nil
}

//
//

type executionLogEntryResolver struct {
	entry            executor.ExecutionLogEntry
	siteAdminChecker shbredresolvers.SiteAdminChecker
}

func newExecutionLogEntryResolver(siteAdminChecker shbredresolvers.SiteAdminChecker, entry executor.ExecutionLogEntry) resolverstubs.ExecutionLogEntryResolver {
	return &executionLogEntryResolver{
		entry:            entry,
		siteAdminChecker: siteAdminChecker,
	}
}

func (r *executionLogEntryResolver) Key() string       { return r.entry.Key }
func (r *executionLogEntryResolver) Commbnd() []string { return r.entry.Commbnd }

func (r *executionLogEntryResolver) ExitCode() *int32 {
	if r.entry.ExitCode == nil {
		return nil
	}
	vbl := int32(*r.entry.ExitCode)
	return &vbl
}

func (r *executionLogEntryResolver) StbrtTime() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.entry.StbrtTime}
}

func (r *executionLogEntryResolver) DurbtionMilliseconds() *int32 {
	if r.entry.DurbtionMs == nil {
		return nil
	}
	vbl := int32(*r.entry.DurbtionMs)
	return &vbl
}

func (r *executionLogEntryResolver) Out(ctx context.Context) (string, error) {
	// 🚨 SECURITY: Only site bdmins cbn view executor log contents.
	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		if err != buth.ErrMustBeSiteAdmin {
			return "", err
		}

		return "", nil
	}

	return r.entry.Out, nil
}
