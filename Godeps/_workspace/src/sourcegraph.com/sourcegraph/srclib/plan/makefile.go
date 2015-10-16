package plan

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"strings"

	"sort"

	"sourcegraph.com/sourcegraph/makex"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/config"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

type Options struct {
	ToolchainExecOpt string

	// When NoCache is true, all files are rebuilt instead of only
	// the ones associated with changed source units.
	NoCache bool

	Verbose bool
}

type RuleMaker func(c *config.Tree, dataDir string, existing []makex.Rule, opt Options) ([]makex.Rule, error)

var (
	RuleMakers        = make(map[string]RuleMaker)
	ruleMakerNames    []string
	orderedRuleMakers []RuleMaker
)

// RegisterRuleMaker adds a function that creates a list of build rules for a
// repository. If RegisterRuleMaker is called twice with the same target or
// target name, if name is empty, or if r is nil, it panics.
func RegisterRuleMaker(name string, r RuleMaker) {
	if _, dup := RuleMakers[name]; dup {
		panic("build: Register called twice for target lister " + name)
	}
	if r == nil {
		panic("build: Register target is nil")
	}
	RuleMakers[name] = r
	ruleMakerNames = append(ruleMakerNames, name)
	orderedRuleMakers = append(orderedRuleMakers, r)
}

// cachedRule is a rule creates the target as a copy of cachedPath. It is
// meant for files that haven't changed between commits.
type cachedRule struct {
	cachedPath string
	target     string
	unit       *unit.SourceUnit
	prereqs    []string
}

func (r *cachedRule) Target() string {
	return r.target
}

func (r *cachedRule) Prereqs() []string {
	return r.prereqs
}

func (r *cachedRule) Recipes() []string {
	return []string{
		// The recipe uses 'cp' instead of 'ln -s' to make it more
		// resilient to things going wrong (like missing the file at
		// cachedPath).
		fmt.Sprintf("cp %s %s", r.cachedPath, r.target),
	}
}

func (r *cachedRule) SourceUnit() *unit.SourceUnit {
	return r.unit
}

// listLatestCommitIDs lists the latest commit ids.
func listLatestCommitIDs(vcsType string) ([]string, error) {
	if vcsType != "git" {
		return nil, fmt.Errorf("listLatestCommitIDs: unsupported vcs type: %q", vcsType)
	}
	cmd := exec.Command("git", "rev-list", "--max-count=5", "HEAD") // 5 picked by random dice roll.
	out, err := cmd.CombinedOutput()
	return strings.Split(string(bytes.TrimSpace(out)), "\n"), err
}

// filesChangedFromRevToIndex returns a list of the files that have
// changed from fromRev to the current index.
func filesChangedFromRevToIndex(vcsType, fromRev string) ([]string, error) {
	if vcsType != "git" {
		return nil, fmt.Errorf("filesChangedFromRevToIndex: unsupported vcs type: %q", vcsType)
	}
	cmd := exec.Command("git", "diff", "--name-only", fromRev, "--")
	out, err := cmd.CombinedOutput()
	return strings.Split(string(bytes.TrimSpace(out)), "\n"), err
}

// CreateMakefile creates the makefiles for the source units in c.
func CreateMakefile(buildDataDir string, buildStore buildstore.RepoBuildStore, vcsType string, c *config.Tree, opt Options) (*makex.Makefile, error) {
	var allRules []makex.Rule
	for i, r := range orderedRuleMakers {
		name := ruleMakerNames[i]
		rules, err := r(c, buildDataDir, allRules, opt)
		if err != nil {
			return nil, fmt.Errorf("rule maker %s: %s", name, err)
		}
		sort.Sort(ruleSort{rules})
		if opt.Verbose {
			log.Printf("%v: Created %d rule(s)", name, len(rules))
		}
		if !opt.NoCache {
			// When cached builds are enabled, we replace all rules whose source unit
			// hasn't changed between commits with a rule that copies the build
			// files stored at the previous commit to the current one.

			// Check to see if a previous build exists.
			var prevCommitID string
			var changedFiles []string
			if revs, err := listLatestCommitIDs(vcsType); err != nil {
				log.Printf("Warning: could not list revisions, rebuilding from scratch: %s, %s", revs, err)
			} else {
				// Skip HEAD, the first revision in the list.
				for i := 1; i < len(revs); i++ {
					if exist, _ := buildstore.BuildDataExistsForCommit(buildStore, revs[i]); !exist {
						continue
					}
					// A build store exists for this commit. Now we need
					// to get all the changed files between this rev and
					// the current rev.
					files, err := filesChangedFromRevToIndex(vcsType, revs[i])
					if err != nil {
						log.Printf("Warning: could not retrieve changed files, rebuilding from scratch: %s %s", files, err)
						break
					}
					changedFiles = files
					prevCommitID = revs[i]
				}
			}
			if prevCommitID != "" {
				// Replace rules.
				for i, rule := range rules {
					r, ok := rule.(interface {
						SourceUnit() *unit.SourceUnit
					})
					if !ok {
						continue
					}
					u := r.SourceUnit()
					if u.ContainsAny(changedFiles) {
						continue
					}

					// The format for p varies based on whether it's prefixed by buildDataDir:
					// if it is, we simply swap the revision in the file name with the
					// previous valid revision. If it isn't, we prefix p with
					// "../[previous-revision]".
					p := strings.Split(rule.Target(), string(filepath.Separator))
					if len(p) > 2 &&
						filepath.Join(p[0:2]...) == buildDataDir &&
						len(p[1]) == 40 { // HACK: Mercurial and Git both use 40-char hashes.
						// p is prefixed by "data-dir/vcs-commit-id"
						p[1] = prevCommitID
					} else {
						p = append([]string{"..", prevCommitID}, p...)
					}

					rules[i] = &cachedRule{
						cachedPath: filepath.Join(p...),
						target:     rule.Target(),
						unit:       u,
						prereqs:    rule.Prereqs(),
					}
				}
			}
		}
		allRules = append(allRules, rules...)
	}

	// Add an "all" rule at the very beginning.
	allTargets := make([]string, len(allRules))
	for i, rule := range allRules {
		allTargets[i] = rule.Target()
	}
	allRule := &makex.BasicRule{TargetFile: "all", PrereqFiles: allTargets}
	allRules = append([]makex.Rule{allRule}, allRules...)

	// DELETE_ON_ERROR makes it so that the targets for failed recipes are
	// deleted. This lets us do "1> $@" to write to the target file without
	// erroneously satisfying the target if the recipe fails. makex has this
	// behavior by default and does not heed .DELETE_ON_ERROR.
	allRules = append(allRules, &makex.BasicRule{TargetFile: ".DELETE_ON_ERROR"})

	mf := &makex.Makefile{Rules: allRules}

	return mf, nil
}
