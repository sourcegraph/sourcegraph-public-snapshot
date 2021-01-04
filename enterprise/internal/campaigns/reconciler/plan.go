package reconciler

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

var operationPrecedence = map[campaigns.ReconcilerOperation]int{
	campaigns.ReconcilerOperationPush:         0,
	campaigns.ReconcilerOperationImport:       1,
	campaigns.ReconcilerOperationPublish:      1,
	campaigns.ReconcilerOperationPublishDraft: 1,
	campaigns.ReconcilerOperationClose:        1,
	campaigns.ReconcilerOperationReopen:       2,
	campaigns.ReconcilerOperationUndraft:      3,
	campaigns.ReconcilerOperationUpdate:       4,
	campaigns.ReconcilerOperationSleep:        5,
	campaigns.ReconcilerOperationSync:         6,
}

type Operations []campaigns.ReconcilerOperation

func (ops Operations) IsNone() bool {
	return len(ops) == 0
}

func (ops Operations) Equal(b Operations) bool {
	if len(ops) != len(b) {
		return false
	}
	bEntries := make(map[campaigns.ReconcilerOperation]struct{})
	for _, e := range b {
		bEntries[e] = struct{}{}
	}

	for _, op := range ops {
		if _, ok := bEntries[op]; !ok {
			return false
		}
	}

	return true
}

func (ops Operations) String() string {
	if ops.IsNone() {
		return "No operations required"
	}
	eo := ops.ExecutionOrder()
	ss := make([]string, len(eo))
	for i, val := range eo {
		ss[i] = strings.ToLower(string(val))
	}
	return strings.Join(ss, " => ")
}

func (ops Operations) ExecutionOrder() []campaigns.ReconcilerOperation {
	uniqueOps := []campaigns.ReconcilerOperation{}

	// Make sure ops are unique.
	seenOps := make(map[campaigns.ReconcilerOperation]struct{})
	for _, op := range ops {
		if _, ok := seenOps[op]; ok {
			continue
		}

		seenOps[op] = struct{}{}
		uniqueOps = append(uniqueOps, op)
	}

	sort.Slice(uniqueOps, func(i, j int) bool {
		return operationPrecedence[uniqueOps[i]] < operationPrecedence[uniqueOps[j]]
	})

	return uniqueOps
}

// Plan represents the possible operations the reconciler needs to do
// to reconcile the current and the desired state of a changeset.
type Plan struct {
	// The changeset that is targeted in this plan.
	Changeset *campaigns.Changeset

	// The changeset spec that is used in this plan.
	ChangesetSpec *campaigns.ChangesetSpec

	// The operations that need to be done to reconcile the changeset.
	Ops Operations

	// The Delta between a possible previous ChangesetSpec and the current
	// ChangesetSpec.
	Delta *ChangesetSpecDelta
}

func (p *Plan) AddOp(op campaigns.ReconcilerOperation) { p.Ops = append(p.Ops, op) }
func (p *Plan) SetOp(op campaigns.ReconcilerOperation) { p.Ops = Operations{op} }

// DeterminePlan looks at the given changeset to determine what action the
// reconciler should take.
// It loads the current ChangesetSpec and if it exists also the previous one.
// If the current ChangesetSpec is not applied to a campaign, it returns an
// error.
func DeterminePlan(previousSpec, currentSpec *campaigns.ChangesetSpec, ch *campaigns.Changeset) (*Plan, error) {
	pl := &Plan{
		Changeset:     ch,
		ChangesetSpec: currentSpec,
	}

	// If it doesn't have a spec, it's an imported changeset and we can't do
	// anything.
	if currentSpec == nil {
		if ch.Unpublished() {
			pl.SetOp(campaigns.ReconcilerOperationImport)
		}
		return pl, nil
	}

	// If it's marked as closing, we don't need to look at the specs.
	if ch.Closing {
		pl.SetOp(campaigns.ReconcilerOperationClose)
		return pl, nil
	}

	delta, err := compareChangesetSpecs(previousSpec, currentSpec)
	if err != nil {
		return pl, nil
	}
	pl.Delta = delta

	switch ch.PublicationState {
	case campaigns.ChangesetPublicationStateUnpublished:
		if currentSpec.Spec.Published.True() {
			pl.SetOp(campaigns.ReconcilerOperationPublish)
			pl.AddOp(campaigns.ReconcilerOperationPush)
		} else if currentSpec.Spec.Published.Draft() && ch.SupportsDraft() {
			// If configured to be opened as draft, and the changeset supports
			// draft mode, publish as draft. Otherwise, take no action.
			pl.SetOp(campaigns.ReconcilerOperationPublishDraft)
			pl.AddOp(campaigns.ReconcilerOperationPush)
		}

	case campaigns.ChangesetPublicationStatePublished:
		// Don't take any actions for merged changesets.
		if ch.ExternalState == campaigns.ChangesetExternalStateMerged {
			return pl, nil
		}
		if reopenAfterDetach(ch) {
			pl.SetOp(campaigns.ReconcilerOperationReopen)
		}

		// Only do undraft, when the codehost supports draft changesets.
		if delta.Undraft && campaigns.ExternalServiceSupports(ch.ExternalServiceType, campaigns.CodehostCapabilityDraftChangesets) {
			pl.AddOp(campaigns.ReconcilerOperationUndraft)
		}

		if delta.AttributesChanged() {
			if delta.NeedCommitUpdate() {
				pl.AddOp(campaigns.ReconcilerOperationPush)
			}

			// If we only need to update the diff and we didn't change the state of the changeset,
			// we're done, because we already pushed the commit. We don't need to
			// update anything on the codehost.
			if !delta.NeedCodeHostUpdate() {
				// But we need to sync the changeset so that it has the new commit.
				//
				// The problem: the code host might not have updated the changeset to
				// have the new commit SHA as its head ref oid (and the check states,
				// ...).
				//
				// That's why we give them 3 seconds to update the changesets.
				//
				// Why 3 seconds? Well... 1 or 2 seem to be too short and 4 too long?
				pl.AddOp(campaigns.ReconcilerOperationSleep)
				pl.AddOp(campaigns.ReconcilerOperationSync)
			} else {
				// Otherwise, we need to update the pull request on the code host or, if we
				// need to reopen it, update it to make sure it has the newest state.
				pl.AddOp(campaigns.ReconcilerOperationUpdate)
			}
		}

	default:
		return pl, fmt.Errorf("unknown changeset publication state: %s", ch.PublicationState)
	}

	return pl, nil
}

func reopenAfterDetach(ch *campaigns.Changeset) bool {
	closed := ch.ExternalState == campaigns.ChangesetExternalStateClosed
	if !closed {
		return false
	}

	// Sanity check: if it's not owned by a campaign, it's simply being tracked.
	if ch.OwnedByCampaignID == 0 {
		return false
	}
	// Sanity check 2: if it's marked as to-be-closed, then we don't reopen it.
	if ch.Closing {
		return false
	}

	// At this point the changeset is closed and not marked as to-be-closed.

	// TODO: What if somebody closed the changeset on purpose on the codehost?
	return ch.AttachedTo(ch.OwnedByCampaignID)
}

func compareChangesetSpecs(previous, current *campaigns.ChangesetSpec) (*ChangesetSpecDelta, error) {
	delta := &ChangesetSpecDelta{}

	if previous == nil {
		return delta, nil
	}

	if previous.Spec.Title != current.Spec.Title {
		delta.TitleChanged = true
	}
	if previous.Spec.Body != current.Spec.Body {
		delta.BodyChanged = true
	}
	if previous.Spec.BaseRef != current.Spec.BaseRef {
		delta.BaseRefChanged = true
	}

	// If was set to "draft" and now "true", need to undraft the changeset.
	// We currently ignore going from "true" to "draft".
	if previous.Spec.Published.Draft() && current.Spec.Published.True() {
		delta.Undraft = true
	}

	// Diff
	currentDiff, err := current.Spec.Diff()
	if err != nil {
		return nil, nil
	}
	previousDiff, err := previous.Spec.Diff()
	if err != nil {
		return nil, err
	}
	if previousDiff != currentDiff {
		delta.DiffChanged = true
	}

	// CommitMessage
	currentCommitMessage, err := current.Spec.CommitMessage()
	if err != nil {
		return nil, nil
	}
	previousCommitMessage, err := previous.Spec.CommitMessage()
	if err != nil {
		return nil, err
	}
	if previousCommitMessage != currentCommitMessage {
		delta.CommitMessageChanged = true
	}

	// AuthorName
	currentAuthorName, err := current.Spec.AuthorName()
	if err != nil {
		return nil, nil
	}
	previousAuthorName, err := previous.Spec.AuthorName()
	if err != nil {
		return nil, err
	}
	if previousAuthorName != currentAuthorName {
		delta.AuthorNameChanged = true
	}

	// AuthorEmail
	currentAuthorEmail, err := current.Spec.AuthorEmail()
	if err != nil {
		return nil, nil
	}
	previousAuthorEmail, err := previous.Spec.AuthorEmail()
	if err != nil {
		return nil, err
	}
	if previousAuthorEmail != currentAuthorEmail {
		delta.AuthorEmailChanged = true
	}

	return delta, nil
}

type ChangesetSpecDelta struct {
	TitleChanged         bool
	BodyChanged          bool
	Undraft              bool
	BaseRefChanged       bool
	DiffChanged          bool
	CommitMessageChanged bool
	AuthorNameChanged    bool
	AuthorEmailChanged   bool
}

func (d *ChangesetSpecDelta) String() string { return fmt.Sprintf("%#v", d) }

func (d *ChangesetSpecDelta) NeedCommitUpdate() bool {
	return d.DiffChanged || d.CommitMessageChanged || d.AuthorNameChanged || d.AuthorEmailChanged
}

func (d *ChangesetSpecDelta) NeedCodeHostUpdate() bool {
	return d.TitleChanged || d.BodyChanged || d.BaseRefChanged
}

func (d *ChangesetSpecDelta) AttributesChanged() bool {
	return d.NeedCommitUpdate() || d.NeedCodeHostUpdate()
}
