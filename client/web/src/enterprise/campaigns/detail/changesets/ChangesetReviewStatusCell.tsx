import React from 'react'
import { ExternalChangesetFields, ChangesetReviewState } from '../../../../graphql-operations'
import DeltaIcon from 'mdi-react/DeltaIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import GateArrowRightIcon from 'mdi-react/GateArrowRightIcon'
import CommentOutlineIcon from 'mdi-react/CommentOutlineIcon'

export interface ChangesetReviewStatusCellProps {
    className?: string
    reviewState: NonNullable<ExternalChangesetFields['reviewState']>
}

export const ChangesetReviewStatusCell: React.FunctionComponent<ChangesetReviewStatusCellProps> = ({ reviewState }) => {
    switch (reviewState) {
        case ChangesetReviewState.APPROVED:
            return <ChangesetReviewStatusApproved />
        case ChangesetReviewState.CHANGES_REQUESTED:
            return <ChangesetReviewStatusChangesRequested />
        case ChangesetReviewState.COMMENTED:
            return <ChangesetReviewStatusCommented />
        case ChangesetReviewState.DISMISSED:
            return <ChangesetReviewStatusDismissed />
        case ChangesetReviewState.PENDING:
            return <ChangesetReviewStatusPending />
    }
}

export const ChangesetReviewStatusPending: React.FunctionComponent<{}> = () => (
    <div className="text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <TimerSandIcon />
        <span className="text-muted">Pending</span>
    </div>
)
export const ChangesetReviewStatusDismissed: React.FunctionComponent<{}> = () => (
    <div className="text-muted m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <GateArrowRightIcon />
        <span className="text-muted">Dismissed</span>
    </div>
)
export const ChangesetReviewStatusCommented: React.FunctionComponent<{}> = () => (
    <div className="text-muted m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <CommentOutlineIcon />
        <span className="text-muted">Commented</span>
    </div>
)
export const ChangesetReviewStatusChangesRequested: React.FunctionComponent<{}> = () => (
    <div className="text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <DeltaIcon />
        <span className="text-muted">Changes requested</span>
    </div>
)
export const ChangesetReviewStatusApproved: React.FunctionComponent<{}> = () => (
    <div className="text-success m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <CheckCircleIcon />
        <span className="text-muted">Approved</span>
    </div>
)
