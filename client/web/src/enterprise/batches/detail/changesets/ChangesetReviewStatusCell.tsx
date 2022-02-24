import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CommentOutlineIcon from 'mdi-react/CommentOutlineIcon'
import DeltaIcon from 'mdi-react/DeltaIcon'
import GateArrowRightIcon from 'mdi-react/GateArrowRightIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

import { ExternalChangesetFields, ChangesetReviewState } from '../../../../graphql-operations'

export interface ChangesetReviewStatusCellProps {
    className?: string
    reviewState: NonNullable<ExternalChangesetFields['reviewState']>
}

export const ChangesetReviewStatusCell: React.FunctionComponent<ChangesetReviewStatusCellProps> = ({
    className,
    reviewState,
}) => {
    switch (reviewState) {
        case ChangesetReviewState.APPROVED:
            return <Icon inline={false} as={ChangesetReviewStatusApproved} className={className} />
        case ChangesetReviewState.CHANGES_REQUESTED:
            return <Icon inline={false} as={ChangesetReviewStatusChangesRequested} className={className} />
        case ChangesetReviewState.COMMENTED:
            return <Icon inline={false} as={ChangesetReviewStatusCommented} className={className} />
        case ChangesetReviewState.DISMISSED:
            return <Icon as={ChangesetReviewStatusDismissed} inline={false} className={className} />
        case ChangesetReviewState.PENDING:
            return <Icon inline={false} as={ChangesetReviewStatusPending} className={className} />
    }
}

export const ChangesetReviewStatusPending: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div
        className={classNames(
            'text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Icon inline={false} as={TimerSandIcon} />
        <span className="text-muted">Pending</span>
    </div>
)
export const ChangesetReviewStatusDismissed: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div
        className={classNames(
            'text-muted m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Icon inline={false} as={GateArrowRightIcon} />
        <span className="text-muted">Dismissed</span>
    </div>
)
export const ChangesetReviewStatusCommented: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div
        className={classNames(
            'text-muted m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Icon inline={false} as={CommentOutlineIcon} />
        <span className="text-muted">Commented</span>
    </div>
)
export const ChangesetReviewStatusChangesRequested: React.FunctionComponent<{ className?: string }> = ({
    className,
}) => (
    <div
        className={classNames(
            'text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Icon inline={false} as={DeltaIcon} />
        <span className="text-muted">Changes requested</span>
    </div>
)
export const ChangesetReviewStatusApproved: React.FunctionComponent<{ className?: string }> = ({ className }) => (
    <div
        className={classNames(
            'text-success m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Icon as={CheckCircleIcon} inline={false} />
        <span className="text-muted">Approved</span>
    </div>
)
