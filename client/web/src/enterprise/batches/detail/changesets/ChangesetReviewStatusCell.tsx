import React from 'react'

import { mdiTimerSand, mdiGateArrowRight, mdiCommentOutline, mdiDelta, mdiCheckCircle } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

import { type ExternalChangesetFields, ChangesetReviewState } from '../../../../graphql-operations'

export interface ChangesetReviewStatusCellProps {
    className?: string
    reviewState: NonNullable<ExternalChangesetFields['reviewState']>
}

export const ChangesetReviewStatusCell: React.FunctionComponent<
    React.PropsWithChildren<ChangesetReviewStatusCellProps>
> = ({ className, reviewState }) => {
    switch (reviewState) {
        case ChangesetReviewState.APPROVED: {
            return <ChangesetReviewStatusApproved className={className} />
        }
        case ChangesetReviewState.CHANGES_REQUESTED: {
            return <ChangesetReviewStatusChangesRequested className={className} />
        }
        case ChangesetReviewState.COMMENTED: {
            return <ChangesetReviewStatusCommented className={className} />
        }
        case ChangesetReviewState.DISMISSED: {
            return <ChangesetReviewStatusDismissed className={className} />
        }
        case ChangesetReviewState.PENDING: {
            return <ChangesetReviewStatusPending className={className} />
        }
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
        <Tooltip content="A review for this changeset is still pending">
            <Icon svgPath={mdiTimerSand} aria-label="A review for this changeset is still pending" inline={false} />
        </Tooltip>
        <span aria-hidden={true} className="text-muted">
            Pending
        </span>
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
        <Icon svgPath={mdiGateArrowRight} aria-label="This changeset's review has been dismissed" inline={false} />
        <span aria-hidden={true} className="text-muted">
            Dismissed
        </span>
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
        <Tooltip content="Comments have been left on this changeset by a reviewer">
            <Icon
                svgPath={mdiCommentOutline}
                aria-label="Comments have been left on this changeset by a reviewer"
                inline={false}
            />
        </Tooltip>
        <span aria-hidden={true} className="text-muted">
            Commented
        </span>
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
        <Icon svgPath={mdiDelta} aria-label="Changes have been requested by a reviewer" inline={false} />
        <span aria-hidden={true} className="text-muted">
            Changes requested
        </span>
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
        <Tooltip content="This changeset has been approved by a reviewer">
            <Icon svgPath={mdiCheckCircle} aria-label="This changeset has been approved by a reviewer" inline={false} />
        </Tooltip>
        <span aria-hidden={true} className="text-muted">
            Approved
        </span>
    </div>
)
