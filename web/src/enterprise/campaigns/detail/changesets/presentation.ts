import { ChangesetState, ChangesetReviewState } from '../../../../../../shared/src/graphql/schema'
import { MdiReactIconComponentType } from 'mdi-react'
import AccountCheckIcon from 'mdi-react/AccountCheckIcon'
import AccountAlertIcon from 'mdi-react/AccountAlertIcon'
import AccountQuestionIcon from 'mdi-react/AccountQuestionIcon'

export const changesetStatusColorClasses: Record<ChangesetState, string> = {
    [ChangesetState.OPEN]: 'success',
    [ChangesetState.CLOSED]: 'danger',
    [ChangesetState.MERGED]: 'purple',
}

export const changesetReviewStateColors: Record<ChangesetReviewState, string> = {
    [ChangesetReviewState.APPROVED]: 'success',
    [ChangesetReviewState.CHANGES_REQUESTED]: 'danger',
    [ChangesetReviewState.PENDING]: 'warning',
}

export const changesetReviewStateIcons: Record<ChangesetReviewState, MdiReactIconComponentType> = {
    [ChangesetReviewState.APPROVED]: AccountCheckIcon,
    [ChangesetReviewState.CHANGES_REQUESTED]: AccountAlertIcon,
    [ChangesetReviewState.PENDING]: AccountQuestionIcon,
}

export const changesetStageLabels: Record<ChangesetReviewState | ChangesetState, string> = {
    [ChangesetState.OPEN]: 'open',
    [ChangesetState.CLOSED]: 'closed',
    [ChangesetState.MERGED]: 'merged',
    [ChangesetReviewState.APPROVED]: 'approved',
    [ChangesetReviewState.CHANGES_REQUESTED]: 'changes requested',
    [ChangesetReviewState.PENDING]: 'pending review',
}
