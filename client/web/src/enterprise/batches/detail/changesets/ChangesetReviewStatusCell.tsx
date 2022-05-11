import React from 'react'

import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CommentOutlineIcon from 'mdi-react/CommentOutlineIcon'
import DeltaIcon from 'mdi-react/DeltaIcon'
import GateArrowRightIcon from 'mdi-react/GateArrowRightIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { ExternalChangesetFields, ChangesetReviewState } from '../../../../graphql-operations'

export interface ChangesetReviewStatusCellProps {
    className?: string
    reviewState: NonNullable<ExternalChangesetFields['reviewState']>
}

export const ChangesetReviewStatusCell: React.FunctionComponent<
    React.PropsWithChildren<ChangesetReviewStatusCellProps>
> = ({ className, reviewState }) => {
    switch (reviewState) {
        case ChangesetReviewState.APPROVED:
            return <ChangesetReviewStatusApproved className={className} />
        case ChangesetReviewState.CHANGES_REQUESTED:
            return <ChangesetReviewStatusChangesRequested className={className} />
        case ChangesetReviewState.COMMENTED:
            return <ChangesetReviewStatusCommented className={className} />
        case ChangesetReviewState.DISMISSED:
            return <ChangesetReviewStatusDismissed className={className} />
        case ChangesetReviewState.PENDING:
            return <ChangesetReviewStatusPending className={className} />
    }
}

export const ChangesetReviewStatusPending: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div
        className={classNames(
            'text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <TimerSandIcon />
        <span className="text-muted">Pending</span>
    </div>
)
export const ChangesetReviewStatusDismissed: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string }>
> = ({ className }) => (
    <div
        className={classNames(
            'text-muted m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <GateArrowRightIcon />
        <span className="text-muted">Dismissed</span>
    </div>
)
export const ChangesetReviewStatusCommented: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string }>
> = ({ className }) => (
    <div
        className={classNames(
            'text-muted m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <CommentOutlineIcon />
        <span className="text-muted">Commented</span>
    </div>
)
export const ChangesetReviewStatusChangesRequested: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string }>
> = ({ className }) => (
    <div
        className={classNames(
            'text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <DeltaIcon />
        <span className="text-muted">Changes requested</span>
    </div>
)
export const ChangesetReviewStatusApproved: React.FunctionComponent<
    React.PropsWithChildren<{ className?: string }>
> = ({ className }) => (
    <div
        className={classNames(
            'text-success m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <CheckCircleIcon />
        <span className="text-muted">Approved</span>
    </div>
)
