import {
    PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields,
    PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields,
    Scalars,
} from '../../../graphql-operations'

/**
 * For a given preview of a changeset to be applied, this method checks if the type of
 * changeset allows for the user to modify its publication state from the UI: namely, that
 * the applied changeset is visible to the user applying, will be attaching or updating
 * the changeset, is not an existing reference, and has not had its publication state set
 * from the batch spec file. Returns the id of the changeset spec if it is publishable
 * from the UI, or null if for any reason it is not.
 *
 * @param node the `ChangesetApplyPreviewFields` node to check
 */
export const getPublishableChangesetSpecID = (
    node:
        | PublishableChangesetSpecIDsVisibleChangesetApplyPreviewFields
        | PublishableChangesetSpecIDsHiddenChangesetApplyPreviewFields
): Scalars['ID'] | null => {
    if (
        node.targets.__typename !== 'VisibleApplyPreviewTargetsAttach' &&
        node.targets.__typename !== 'VisibleApplyPreviewTargetsUpdate'
    ) {
        return null
    }
    if (node.targets.changesetSpec.description.__typename !== 'GitBranchChangesetDescription') {
        return null
    }
    if (node.targets.changesetSpec.description.published !== null) {
        return null
    }
    return node.targets.changesetSpec.id
}
