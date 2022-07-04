import React from 'react'

import { mdiTimerSand, mdiCheckCircle, mdiCloseCircle } from '@mdi/js'
import classNames from 'classnames'

import { Icon } from '@sourcegraph/wildcard'

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
        <Icon data-tooltip="Check state is pending" svgPath={mdiTimerSand} inline={false} aria-hidden={true} />
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
        <Icon data-tooltip="All checks complete" svgPath={mdiCheckCircle} inline={false} aria-hidden={true} />
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
        <Icon data-tooltip="Some checks failed" svgPath={mdiCloseCircle} inline={false} aria-hidden={true} />
        <span className="text-muted">Failed</span>
    </div>
)
