package campaigns

// This file contains methods that exist solely to migrate campaigns and
// changesets lingering from before specs were added in Sourcegraph 3.19 into
// the new world.
//
// It should be removed in or after Sourcegraph 3.21.

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
)

// preSpecChangesetColumns are used by by the changeset related Store methods
// and by workerutil.Worker to load changesets from the database for processing
// by the reconciler.
var preSpecChangesetColumns = []*sqlf.Query{
	sqlf.Sprintf("changesets.id"),
	sqlf.Sprintf("changesets.repo_id"),
	sqlf.Sprintf("changesets.created_at"),
	sqlf.Sprintf("changesets.updated_at"),
	sqlf.Sprintf("changesets.metadata"),
	sqlf.Sprintf("changesets.campaign_ids"),
	sqlf.Sprintf("changesets.external_id"),
	sqlf.Sprintf("changesets.external_service_type"),
	sqlf.Sprintf("changesets.external_branch"),
	sqlf.Sprintf("changesets.external_deleted_at"),
	sqlf.Sprintf("changesets.external_updated_at"),
	sqlf.Sprintf("changesets.external_state"),
	sqlf.Sprintf("changesets.external_review_state"),
	sqlf.Sprintf("changesets.external_check_state"),
	sqlf.Sprintf("changesets.created_by_campaign"),
	sqlf.Sprintf("changesets.added_to_campaign"),
	sqlf.Sprintf("changesets.diff_stat_added"),
	sqlf.Sprintf("changesets.diff_stat_changed"),
	sqlf.Sprintf("changesets.diff_stat_deleted"),
	sqlf.Sprintf("changesets.sync_state"),
	sqlf.Sprintf("changesets.owned_by_campaign_id"),
	sqlf.Sprintf("changesets.current_spec_id"),
	sqlf.Sprintf("changesets.previous_spec_id"),
	sqlf.Sprintf("changesets.publication_state"),
	sqlf.Sprintf("changesets.reconciler_state"),
	sqlf.Sprintf("changesets.failure_message"),
	sqlf.Sprintf("changesets.started_at"),
	sqlf.Sprintf("changesets.finished_at"),
	sqlf.Sprintf("changesets.process_after"),
	sqlf.Sprintf("changesets.num_resets"),
}

func (s *Store) listPreSpecChangesets(ctx context.Context) ([]*campaigns.Changeset, error) {
	cs := []*campaigns.Changeset{}
	q := sqlf.Sprintf(
		listPreSpecChangesetsFmtstr,
		sqlf.Join(preSpecChangesetColumns, ", "),
	)
	if err := s.query(ctx, q, func(sc scanner) error {
		var c campaigns.Changeset
		if err := scanPreSpecChangeset(&c, sc); err != nil {
			return err
		}
		cs = append(cs, &c)
		return nil
	}); err != nil {
		return nil, err
	}

	return cs, nil
}

const listPreSpecChangesetsFmtstr = `
-- source: enterprise/internal/campaigns/spec_migration.go:listPreSpecChangesets
SELECT %s FROM changesets_old
`

func scanPreSpecChangeset(t *campaigns.Changeset, s scanner) error {
	var metadata, syncState json.RawMessage

	var (
		externalState       string
		externalReviewState string
		externalCheckState  string
		failureMessage      string
		reconcilerState     string
	)
	err := s.Scan(
		&t.ID,
		&t.RepoID,
		&t.CreatedAt,
		&t.UpdatedAt,
		&metadata,
		&dbutil.JSONInt64Set{Set: &t.CampaignIDs},
		&dbutil.NullString{S: &t.ExternalID},
		&t.ExternalServiceType,
		&dbutil.NullString{S: &t.ExternalBranch},
		&dbutil.NullTime{Time: &t.ExternalDeletedAt},
		&dbutil.NullTime{Time: &t.ExternalUpdatedAt},
		&dbutil.NullString{S: &externalState},
		&dbutil.NullString{S: &externalReviewState},
		&dbutil.NullString{S: &externalCheckState},
		&t.CreatedByCampaign,
		&t.AddedToCampaign,
		&t.DiffStatAdded,
		&t.DiffStatChanged,
		&t.DiffStatDeleted,
		&syncState,
		&dbutil.NullInt64{N: &t.OwnedByCampaignID},
		&dbutil.NullInt64{N: &t.CurrentSpecID},
		&dbutil.NullInt64{N: &t.PreviousSpecID},
		&t.PublicationState,
		&reconcilerState,
		&dbutil.NullString{S: &failureMessage},
		&dbutil.NullTime{Time: &t.StartedAt},
		&dbutil.NullTime{Time: &t.FinishedAt},
		&dbutil.NullTime{Time: &t.ProcessAfter},
		&t.NumResets,
	)
	if err != nil {
		return errors.Wrap(err, "scanning changeset")
	}

	t.ExternalState = campaigns.ChangesetExternalState(externalState)
	t.ExternalReviewState = campaigns.ChangesetReviewState(externalReviewState)
	t.ExternalCheckState = campaigns.ChangesetCheckState(externalCheckState)
	if failureMessage != "" {
		t.FailureMessage = &failureMessage
	}
	t.ReconcilerState = campaigns.ReconcilerState(strings.ToUpper(reconcilerState))

	switch t.ExternalServiceType {
	case extsvc.TypeGitHub:
		t.Metadata = new(github.PullRequest)
	case extsvc.TypeBitbucketServer:
		t.Metadata = new(bitbucketserver.PullRequest)
	case extsvc.TypeGitLab:
		t.Metadata = new(gitlab.MergeRequest)
	default:
		return errors.New("unknown external service type")
	}

	if err = json.Unmarshal(metadata, t.Metadata); err != nil {
		return errors.Wrapf(err, "scanChangeset: failed to unmarshal %q metadata", t.ExternalServiceType)
	}
	if err = json.Unmarshal(syncState, &t.SyncState); err != nil {
		return errors.Wrapf(err, "scanChangeset: failed to unmarshal sync state: %s", syncState)
	}

	return nil
}
