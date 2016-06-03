package localstore

import (
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/coverageutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/ext/slack"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	AppSchema.Map.AddTableWithName(dbRepoStatus{}, "repo_status").SetKeys(false, "Repo", "Rev", "Context")
	AppSchema.CreateSQL = append(AppSchema.CreateSQL,
		`ALTER TABLE repo_status ALTER COLUMN repo TYPE citext;`,
		`ALTER TABLE repo_status ALTER COLUMN description TYPE text;`,
		`ALTER TABLE repo_status ALTER COLUMN created_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
		`ALTER TABLE repo_status ALTER COLUMN updated_at TYPE timestamp with time zone USING updated_at::timestamp with time zone;`,
	)
}

// dbRepoStatus DB-maps a sourcegraph.RepoStatus object.
type dbRepoStatus struct {
	Repo        string
	Rev         string
	State       string
	TargetURL   string `db:"target_url"`
	Description string
	Context     string
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at"`
}

type dbFileCoverage struct {
	Path             string
	Idents           int
	Refs             int
	Defs             int
	UnresolvedIdents []*coverageutil.Token
	UnresolvedRefs   []*coverageutil.Token
}

type dbRepoCoverage struct {
	Repo          string
	Rev           string
	Language      string
	Files         []*dbFileCoverage
	Summary       *dbFileCoverage
	SrclibVersion string
	Day           string
	Duration      float64
}

func (r *dbRepoStatus) toRepoStatus() *sourcegraph.RepoStatus {
	r2 := &sourcegraph.RepoStatus{
		State:       r.State,
		TargetURL:   r.TargetURL,
		Description: r.Description,
		Context:     r.Context,
		CreatedAt:   pbtypes.NewTimestamp(r.CreatedAt),
	}

	if r.UpdatedAt != nil {
		r2.UpdatedAt = pbtypes.NewTimestamp(*r.UpdatedAt)
	}

	return r2
}

func (r *dbRepoStatus) fromRepoStatus(repoRev *sourcegraph.RepoRevSpec, r2 *sourcegraph.RepoStatus) {
	r.Repo = repoRev.URI
	r.Rev = repoRev.CommitID
	r.State = r2.State
	r.TargetURL = r2.TargetURL
	r.Description = r2.Description
	r.Context = r2.Context
	r.CreatedAt = r2.CreatedAt.Time()
	if !r2.UpdatedAt.Time().IsZero() {
		ts := r2.UpdatedAt.Time()
		r.UpdatedAt = &ts
	}
}

func toRepoStatuses(rs []*dbRepoStatus) []*sourcegraph.RepoStatus {
	r2s := make([]*sourcegraph.RepoStatus, len(rs))
	for i, r := range rs {
		r2s[i] = r.toRepoStatus()
	}
	return r2s
}

type repoStatuses struct{}

var _ store.RepoStatuses = (*repoStatuses)(nil)

func (s *repoStatuses) GetCombined(ctx context.Context, repoRev sourcegraph.RepoRevSpec) (*sourcegraph.CombinedStatus, error) {
	if err := accesscontrol.VerifyUserHasReadAccess(ctx, "RepoStatuses.GetCombined", repoRev.URI); err != nil {
		return nil, err
	}
	rev := repoRev.CommitID

	var dbRepoStatuses []*dbRepoStatus
	if _, err := appDBH(ctx).Select(&dbRepoStatuses, `SELECT * FROM repo_status WHERE repo=$1 AND rev=$2 ORDER BY created_at ASC;`, repoRev.URI, rev); err != nil {
		return nil, err
	}
	return &sourcegraph.CombinedStatus{
		Rev:      repoRev.CommitID,
		CommitID: repoRev.CommitID,
		Statuses: toRepoStatuses(dbRepoStatuses),
	}, nil
}

func (s *repoStatuses) GetCoverage(ctx context.Context) (*sourcegraph.RepoStatusList, error) {
	if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "RepoStatuses.GetCoverage"); err != nil {
		return nil, err
	}

	var dbRepoStatuses []*dbRepoStatus
	if _, err := appDBH(ctx).Select(&dbRepoStatuses, `SELECT * FROM repo_status WHERE context=$1`, "coverage"); err != nil {
		return nil, err
	}

	list := sourcegraph.RepoStatusList{}
	for _, status := range dbRepoStatuses {
		list.RepoStatuses = append(list.RepoStatuses, status.toRepoStatus())
	}
	return &list, nil
}

func checkCoverageRegression(prev, next *dbRepoCoverage) {
	ps := prev.Summary
	ns := next.Summary

	refScore := func(cvg *dbFileCoverage) float64 {
		if cvg == nil || cvg.Idents == 0 {
			return 0
		}
		return float64(ps.Refs) / float64(ps.Idents)
	}
	defScore := func(cvg *dbFileCoverage) float64 {
		if cvg == nil || cvg.Idents == 0 {
			return 0
		}
		return float64(ps.Defs) / float64(ps.Idents)
	}

	if refScore(ps) > refScore(ns) || defScore(ps) > defScore(ns) {
		slack.PostMessage(slack.PostOpts{
			Msg: fmt.Sprintf(`Coverage for %s (lang=%s) has regressed.
Before: refScore(%d), defScore(%d)
After: refScore(%d), defScore(%d)`, prev.Repo, prev.Language, refScore(ps), defScore(ps), refScore(ns), defScore(ns)),
			IconEmoji: ":warning:",
			Channel:   "global-graph",
		})
	}
}

func (s *repoStatuses) Create(ctx context.Context, repoRev sourcegraph.RepoRevSpec, status *sourcegraph.RepoStatus) error {
	if err := accesscontrol.VerifyUserHasWriteAccess(ctx, "RepoStatuses.Create", repoRev.URI); err != nil {
		return err
	}

	var dbrs dbRepoStatus
	dbrs.fromRepoStatus(&repoRev, status)
	if dbrs.CreatedAt.Unix() == 0 {
		dbrs.CreatedAt = time.Now()
	}

	// Upsert the status. Note that this is correct, because repo
	// statuses cannot be deleted. It is more robust to write it this
	// way than with inline SQL (which would have to be manually
	// updated if the fields of RepoStatus changed).
	err := appDBH(ctx).Insert(&dbrs)
	if err != nil {
		if dbrs.Context == "coverage" {
			// Repo coverage is stored as a JSON encoded array of dbRepoCoverage
			// (roughly one per day, not guaranteed). If a row already exists, prepend
			// the next coverage stat to the previous.
			prevStatus := dbRepoStatus{}
			if err := appDBH(ctx).SelectOne(&prevStatus, `SELECT * FROM repo_status WHERE repo=$1 AND rev=$2 AND context=$3;`, repoRev.URI, repoRev.CommitID, "coverage"); err != nil {
				return err
			}

			var cvg []dbRepoCoverage
			if err := json.Unmarshal([]byte(prevStatus.Description), &cvg); err != nil {
				return err
			}

			var nextCvg []dbRepoCoverage // should be length 1
			if err := json.Unmarshal([]byte(dbrs.Description), &nextCvg); err != nil {
				return err
			}
			if len(nextCvg) != 1 {
				return fmt.Errorf("must add one coverage stat at a time per repo") // invariant
			}

			if len(cvg) > 0 { // sanity check
				prev := cvg[0]
				next := nextCvg[0]

				checkCoverageRegression(&prev, &next)

				cvg[0].Files = nil // keep granular per-file data only for the most recent day
				if prev.Day != next.Day {
					cvg = append(nextCvg, cvg...) // keep most recent day first; don't double track days
				} else {
					// overwrite prev day with new stat
					cvg[0] = next
				}
			} else {
				cvg = nextCvg
			}

			if len(cvg) > 7 {
				cvg = cvg[:7] // cap # of entries at 7 days
			}
			nextDescription, err := json.Marshal(&cvg)
			if err != nil {
				return err
			}

			dbrs.Description = string(nextDescription)
		}

		_, err := appDBH(ctx).Update(&dbrs)
		return err
	}
	return err
}
