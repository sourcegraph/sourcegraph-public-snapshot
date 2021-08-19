import {
    ChangesetState,
    PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields,
    PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields,
    Scalars,
} from '../../../graphql-operations'

/**
 * For a given preview of a changeset to be applied, this method checks if the type of
 * changeset allows for the user to modify its publication state from the UI: namely, that
 * all of the following conditions are true:
 * - the changeset is visible to the user applying
 * - the operation will be attaching or updating a changeset
 * - the changeset is not an existing reference
 * - the changeset does not have its publication state specified in the batch spec file
 * - if the operation is updating a changeset, the changeset is in a state we can
 * transition to published or draft from
 *
 * Returns the id of the changeset spec if it is publishable from the UI, or null if for
 * any reason it is not.
 *
 * @param node the `ChangesetApplyPreviewFields` node to check
 */
export const getPublishableChangesetSpecID = (
    node:
        | PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields
        | PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields
): Scalars['ID'] | null => {
    // The changeset is either hidden, or the operation is detaching a changeset
    if (
        node.targets.__typename !== 'VisibleApplyPreviewTargetsAttach' &&
        node.targets.__typename !== 'VisibleApplyPreviewTargetsUpdate'
    ) {
        return null
    }
    // The changeset is an existing reference
    if (node.targets.changesetSpec.description.__typename !== 'GitBranchChangesetDescription') {
        return null
    }
    // The changeset has its publication state specified in the batch spec file, which takes priority
    if (node.targets.changesetSpec.description.published !== null) {
        return null
    }
    // The changeset is already published or in a state we can't transition to published/draft from
    if (
        node.targets.__typename === 'VisibleApplyPreviewTargetsUpdate' &&
        !(node.targets.changeset.state in [ChangesetState.DRAFT, ChangesetState.UNPUBLISHED])
    ) {
        return null
    }
    return node.targets.changesetSpec.id
}
