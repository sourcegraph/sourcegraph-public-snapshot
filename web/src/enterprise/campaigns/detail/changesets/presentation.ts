import {
    ChangesetExternalState,
    ChangesetReviewState,
    ChangesetCheckState,
} from '../../../../../../shared/src/graphql/schema'
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

export const changesetExternalStateColorClasses: Record<ChangesetExternalState, string> = {
    [ChangesetExternalState.OPEN]: 'success',
    [ChangesetExternalState.CLOSED]: 'danger',
    [ChangesetExternalState.DELETED]: 'muted',
    [ChangesetExternalState.MERGED]: 'merged',
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

export const changesetStateLabels: Record<ChangesetReviewState | ChangesetExternalState, string> = {
    [ChangesetExternalState.OPEN]: 'open',
    [ChangesetExternalState.CLOSED]: 'closed',
    [ChangesetExternalState.MERGED]: 'merged',
    [ChangesetExternalState.DELETED]: 'deleted',
    [ChangesetReviewState.APPROVED]: 'approved',
    [ChangesetReviewState.CHANGES_REQUESTED]: 'changes requested',
    [ChangesetReviewState.PENDING]: 'pending review',
    [ChangesetReviewState.COMMENTED]: 'commented',
    [ChangesetReviewState.DISMISSED]: 'dismissed',
}

export const changesetExternalStateIcons: Record<ChangesetExternalState, MdiReactIconComponentType> = {
    [ChangesetExternalState.CLOSED]: SourcePullIcon,
    [ChangesetExternalState.MERGED]: SourceMergeIcon,
    [ChangesetExternalState.OPEN]: SourcePullIcon,
    [ChangesetExternalState.DELETED]: DeleteIcon,
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
