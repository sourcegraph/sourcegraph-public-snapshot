import { ChangesetState, ChangesetReviewState, ChangesetCheckState } from '../../../../../../shared/src/graphql/schema'
import { MdiReactIconComponentType } from 'mdi-react'
import AccountCheckIcon from 'mdi-react/AccountCheckIcon'
import AccountAlertIcon from 'mdi-react/AccountAlertIcon'
import AccountQuestionIcon from 'mdi-react/AccountQuestionIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import CheckboxBlankCircleIcon from 'mdi-react/CheckboxBlankCircleIcon'
import { CampaignsIcon } from '../../icons'
import SyncIcon from 'mdi-react/SyncIcon'

export const changesetStatusColorClasses: Record<ChangesetState, string> = {
    [ChangesetState.OPEN]: 'success',
    [ChangesetState.CLOSED]: 'danger',
    [ChangesetState.DELETED]: 'muted',
    [ChangesetState.MERGED]: 'merged',
    [ChangesetState.PENDING]: 'muted',
    [ChangesetState.PUBLISHING]: 'info',
    [ChangesetState.ERRORED]: 'danger',
}

export const changesetReviewStateColors: Record<ChangesetReviewState, string> = {
    [ChangesetReviewState.APPROVED]: 'success',
    [ChangesetReviewState.CHANGES_REQUESTED]: 'danger',
    [ChangesetReviewState.PENDING]: 'warning',
    [ChangesetReviewState.COMMENTED]: 'warning',
    [ChangesetReviewState.DISMISSED]: 'warning',
}

export const changesetReviewStateIcons: Record<ChangesetReviewState, MdiReactIconComponentType> = {
    [ChangesetReviewState.APPROVED]: AccountCheckIcon,
    [ChangesetReviewState.CHANGES_REQUESTED]: AccountAlertIcon,
    [ChangesetReviewState.PENDING]: AccountQuestionIcon,
    [ChangesetReviewState.COMMENTED]: AccountQuestionIcon,
    [ChangesetReviewState.DISMISSED]: AccountQuestionIcon,
}

export const changesetStateLabels: Record<ChangesetState, string> = {
    [ChangesetState.OPEN]: 'open',
    [ChangesetState.CLOSED]: 'closed',
    [ChangesetState.MERGED]: 'merged',
    [ChangesetState.DELETED]: 'deleted',
    [ChangesetState.PENDING]: 'pending',
    [ChangesetState.PUBLISHING]: 'publishing',
    [ChangesetState.ERRORED]: 'errored',
}

export const changesetReviewStateLabels: Record<ChangesetReviewState, string> = {
    [ChangesetReviewState.APPROVED]: 'approved',
    [ChangesetReviewState.CHANGES_REQUESTED]: 'changes requested',
    [ChangesetReviewState.PENDING]: 'pending review',
    [ChangesetReviewState.COMMENTED]: 'commented',
    [ChangesetReviewState.DISMISSED]: 'dismissed',
}

export const changesetStateIcons: Record<ChangesetState, MdiReactIconComponentType> = {
    [ChangesetState.CLOSED]: SourcePullIcon,
    [ChangesetState.MERGED]: SourceMergeIcon,
    [ChangesetState.OPEN]: SourcePullIcon,
    [ChangesetState.DELETED]: DeleteIcon,
    [ChangesetState.PENDING]: CampaignsIcon,
    [ChangesetState.PUBLISHING]: SyncIcon,
    [ChangesetState.ERRORED]: ErrorIcon,
}

export const changesetCheckStateIcons: Record<ChangesetCheckState, MdiReactIconComponentType> = {
    [ChangesetCheckState.PENDING]: CheckboxBlankCircleIcon,
    [ChangesetCheckState.PASSED]: CheckCircleIcon,
    [ChangesetCheckState.FAILED]: ErrorIcon,
}

export const changesetCheckStateColors: Record<ChangesetCheckState, string> = {
    [ChangesetCheckState.PENDING]: 'text-warning',
    [ChangesetCheckState.PASSED]: 'text-success',
    [ChangesetCheckState.FAILED]: 'text-danger',
}

export const changesetCheckStateTooltips: Record<ChangesetCheckState, string> = {
    [ChangesetCheckState.PENDING]: 'Check state is pending',
    [ChangesetCheckState.PASSED]: 'All checks complete',
    [ChangesetCheckState.FAILED]: 'Some checks failed',
}
