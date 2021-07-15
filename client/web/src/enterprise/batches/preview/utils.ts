import { HiddenChangesetApplyPreviewFields, VisibleChangesetApplyPreviewFields } from '../../../graphql-operations'

export const canSetPublishedState = (
    node: HiddenChangesetApplyPreviewFields | VisibleChangesetApplyPreviewFields
): string | null => {
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
