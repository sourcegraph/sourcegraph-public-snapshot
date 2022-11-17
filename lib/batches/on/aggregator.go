package on

// RepoRevisionAggregator implements the precedence rules used when resolving
// the on: rules in a single batch spec. Specifically, later rules generally
// override earlier rules, but repository: rules always override
// repositoriesMatchingQuery: rules.
//
// This is essentially a generic type, with two parameters (albeit these are
// mostly exposed in OnResult:
//
//   - RepoID: An opaque identifier used to identify unique repositories. This
//     must be able to be used as a map key.
//   - Revision: An object that identifies the specific revision. There are no
//     requirements for this type, as it will be returned as-is in
//     Revisions().
type RepoRevisionAggregator struct {
	results []*RuleRevisions
}

type RepoID any
type Revision any

func NewRepoRevisionAggregator() *RepoRevisionAggregator {
	return &RepoRevisionAggregator{
		results: []*RuleRevisions{},
	}
}

// RepositoryRuleType represents the type of on: rule that was used to create
// the result.
type RepositoryRuleType int

const (
	// RepositoryRuleTypeQuery is a repositoriesMatchingQuery: rule.
	RepositoryRuleTypeQuery RepositoryRuleType = iota

	// RepositoryRuleTypeExplicit is a repository: rule.
	RepositoryRuleTypeExplicit
)

// NewRuleRevisions instantiates a new RuleRevisions, which is used to track the
// revisions returned by a specific on: rule. The rule type must be provided as
// the ruleType argument.
func (agg *RepoRevisionAggregator) NewRuleRevisions(ruleType RepositoryRuleType) *RuleRevisions {
	result := &RuleRevisions{
		ruleType: ruleType,
		repos:    map[RepoID][]Revision{},
	}
	agg.results = append(agg.results, result)

	return result
}

// Revisions returns all the revisions matched by the rules added to the
// aggregator, applying the on: precedence rules as it iterates.
func (agg *RepoRevisionAggregator) Revisions() []Revision {
	type repo struct {
		ruleType  RepositoryRuleType
		revisions []Revision
	}
	repos := map[RepoID]repo{}

	for _, result := range agg.results {
		for id, revisions := range result.repos {
			if previous, ok := repos[id]; ok {
				// The only scenario where the new repo will not replace the
				// previous repo is when the previous repo was the result of an
				// explicit repository: rule and the new repo was the result of
				// a repositoriesMatchingQuery: rule.
				if previous.ruleType == RepositoryRuleTypeExplicit && result.ruleType == RepositoryRuleTypeQuery {
					continue
				}
			}

			repos[id] = repo{
				ruleType:  result.ruleType,
				revisions: revisions,
			}
		}
	}

	revisions := []Revision{}
	for _, r := range repos {
		revisions = append(revisions, r.revisions...)
	}

	return revisions
}

// RuleRevisions is used to capture the revisions added by a single on: rule.
type RuleRevisions struct {
	ruleType RepositoryRuleType
	repos    map[RepoID][]Revision
}

// AddRepoRevision adds a single repo revision to the results of this rule.
func (result *RuleRevisions) AddRepoRevision(repo RepoID, revision Revision) {
	result.repos[repo] = append(result.repos[repo], revision)
}
