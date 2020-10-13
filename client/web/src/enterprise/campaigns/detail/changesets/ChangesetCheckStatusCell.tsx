import React from 'react'
import { ExternalChangesetFields, ChangesetCheckState } from '../../../../graphql-operations'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CloseCircleIcon from 'mdi-react/CloseCircleIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

export interface ChangesetCheckStatusCellProps {
    className?: string
    checkState: NonNullable<ExternalChangesetFields['checkState']>
}

export const ChangesetCheckStatusCell: React.FunctionComponent<ChangesetCheckStatusCellProps> = ({ checkState }) => {
    switch (checkState) {
        case ChangesetCheckState.PENDING:
            return <ChangesetCheckStatusPending />
        case ChangesetCheckState.PASSED:
            return <ChangesetCheckStatusPassed />
        case ChangesetCheckState.FAILED:
            return <ChangesetCheckStatusFailed />
    }
}

export const ChangesetCheckStatusPending: React.FunctionComponent<{}> = () => (
    <div className="text-warning m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <TimerSandIcon data-tooltip="Check state is pending" />
        <span className="text-muted">Pending</span>
    </div>
)
export const ChangesetCheckStatusPassed: React.FunctionComponent<{}> = () => (
    <div className="text-success m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <CheckCircleIcon data-tooltip="All checks complete" />
        <span className="text-muted">Passed</span>
    </div>
)
export const ChangesetCheckStatusFailed: React.FunctionComponent<{}> = () => (
    <div className="text-danger m-0 text-nowrap d-flex flex-column align-items-center justify-content-center">
        <CloseCircleIcon data-tooltip="Some checks failed" />
        <span className="text-muted">Failed</span>
    </div>
)
