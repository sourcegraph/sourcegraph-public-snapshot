import {
    ChangesetState,
    type PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields,
    type PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields,
    type Scalars,
} from '../../../graphql-operations'

/* The preview changeset can be published from the UI */
export interface Publishable {
    publishable: true
    changesetSpecID: Scalars['ID']
}

/* The preview changeset cannot currently be published from the UI */
export interface Unpublishable {
    publishable: false
    reason: string
}

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
 * Returns a `Publishable` with the node's changeset spec ID if it is publishable from the
 * UI, or else an `Unpublishable` with the reason it is not.
 *
 * @param node the `ChangesetApplyPreviewFields` node to check
 */
export const checkPublishability = (
    node:
        | PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields
        | PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields
): Publishable | Unpublishable => {
    // The operation is detaching a changeset
    if (node.targets.__typename === 'VisibleApplyPreviewTargetsDetach') {
        return {
            publishable: false,
            reason: 'You cannot modify the publication state for a changeset that will be detached from this batch change.',
        }
    }
    // The changeset is hidden to the user applying
    if (
        node.targets.__typename !== 'VisibleApplyPreviewTargetsAttach' &&
        node.targets.__typename !== 'VisibleApplyPreviewTargetsUpdate'
    ) {
        return { publishable: false, reason: 'You do not have permission to publish to this repository.' }
    }
    // The changeset is an existing, imported reference
    if (node.targets.changesetSpec.description.__typename !== 'GitBranchChangesetDescription') {
        return {
            publishable: false,
            reason: 'You cannot modify the publication state for an imported changeset.',
        }
    }
    // The changeset has its publication state specified in the batch spec file, which takes priority
    if (node.targets.changesetSpec.description.published !== null) {
        return {
            publishable: false,
            reason: "This changeset's publication state is being controlled by the spec file. To modify it here, omit it from your spec.",
        }
    }
    if (node.targets.__typename === 'VisibleApplyPreviewTargetsUpdate') {
        // The changeset is already published
        if (
            [
                ChangesetState.CLOSED,
                ChangesetState.DELETED,
                ChangesetState.MERGED,
                ChangesetState.OPEN,
                ChangesetState.READONLY,
            ].includes(node.targets.changeset.state)
        ) {
            return { publishable: false, reason: 'This changeset has already been published.' }
        }
        // This changeset is not in a state we want to transition to published/draft from
        if (![ChangesetState.DRAFT, ChangesetState.UNPUBLISHED].includes(node.targets.changeset.state)) {
            return { publishable: false, reason: 'This changeset is in a state we cannot currently publish from.' }
        }
    }
    return { publishable: true, changesetSpecID: node.targets.changesetSpec.id }
}

/**
 * For a list of `ChangesetApplyPreview` nodes, this method returns a list of the
 * changeset spec IDs for any of the nodes that are considered publishable.
 *
 * @see `checkPublishability`
 */
export const filterPublishableIDs = (
    nodes: (
        | PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields
        | PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields
    )[]
): Scalars['ID'][] =>
    nodes
        .map(node => checkPublishability(node))
        .filter((result): result is Publishable => result.publishable)
        .map(result => result.changesetSpecID)
