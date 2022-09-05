import React from 'react'

import { mdiCheckCircle, mdiCloseCircle, mdiTimerSand } from '@mdi/js'
import classNames from 'classnames'

import { Icon, Tooltip } from '@sourcegraph/wildcard'

import { ExternalChangesetFields, ChangesetCheckState } from '../../../../graphql-operations'

export interface ChangesetCheckStatusCellProps {
    className?: string
    checkState: NonNullable<ExternalChangesetFields['checkState']>
}

export const ChangesetCheckStatusCell: React.FunctionComponent<
    React.PropsWithChildren<ChangesetCheckStatusCellProps>
> = ({ className, checkState }) => {
    switch (checkState) {
        case ChangesetCheckState.PENDING:
            return <ChangesetCheckStatusPending className={className} />
        case ChangesetCheckState.PASSED:
            return <ChangesetCheckStatusPassed className={className} />
        case ChangesetCheckState.FAILED:
            return <ChangesetCheckStatusFailed className={className} />
    }
}

export const ChangesetCheckStatusPending: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div
        className={classNames(
            'text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Tooltip content="Check state is pending">
            <Icon svgPath={mdiTimerSand} aria-label="Check state is pending" inline={false} />
        </Tooltip>
        <span className="text-muted">Pending</span>
    </div>
)
export const ChangesetCheckStatusPassed: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div
        className={classNames(
            'text-success m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Tooltip content="All checks complete">
            <Icon svgPath={mdiCheckCircle} aria-label="All checks complete" inline={false} />
        </Tooltip>
        <span className="text-muted">Passed</span>
    </div>
)
export const ChangesetCheckStatusFailed: React.FunctionComponent<React.PropsWithChildren<{ className?: string }>> = ({
    className,
}) => (
    <div
        className={classNames(
            'text-danger m-0 text-nowrap d-flex flex-column align-items-center justify-content-center',
            className
        )}
    >
        <Tooltip content="Some checks failed">
            <Icon svgPath={mdiCloseCircle} aria-label="Some checks failed" inline={false} />
        </Tooltip>
        <span className="text-muted">Failed</span>
    </div>
)
