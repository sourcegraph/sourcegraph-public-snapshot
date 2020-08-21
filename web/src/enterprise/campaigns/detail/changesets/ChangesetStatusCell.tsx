import React from 'react'
import { ChangesetFields } from '../../../../graphql-operations'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import classNames from 'classnames'
import { computeChangesetUIState, ChangesetUIState } from '../../utils'

export interface ChangesetStatusCellProps {
    className?: string
    changeset: Pick<ChangesetFields, 'publicationState' | 'externalState' | 'reconcilerState'>
}

export const ChangesetStatusCell: React.FunctionComponent<ChangesetStatusCellProps> = ({ changeset }) => {
    switch (computeChangesetUIState(changeset)) {
        case ChangesetUIState.ERRORED:
            return <ChangesetStatusError />
        case ChangesetUIState.PROCESSING:
            return <ChangesetStatusProcessing />
        case ChangesetUIState.UNPUBLISHED:
            return <ChangesetStatusUnpublished />
        case ChangesetUIState.OPEN:
            return <ChangesetStatusOpen />
        case ChangesetUIState.CLOSED:
            return <ChangesetStatusClosed />
        case ChangesetUIState.MERGED:
            return <ChangesetStatusMerged />
        case ChangesetUIState.DELETED:
            return <ChangesetStatusDeleted />
    }
}

const iconClassNames = 'm-0 text-nowrap d-flex flex-column align-items-center justify-content-center'

export const ChangesetStatusUnpublished: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Unpublished</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-muted', className)}>
        <SourceBranchIcon />
        {label}
    </div>
)
export const ChangesetStatusClosed: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Closed</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-danger', className)}>
        <SourcePullIcon />
        {label}
    </div>
)
export const ChangesetStatusMerged: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Merged</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-merged', className)}>
        <SourceMergeIcon />
        {label}
    </div>
)
export const ChangesetStatusOpen: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Open</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-success', className)}>
        <SourcePullIcon />
        {label}
    </div>
)
export const ChangesetStatusDeleted: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Deleted</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-muted', className)}>
        <DeleteIcon />
        {label}
    </div>
)
export const ChangesetStatusError: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-danger">Error</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-danger', className)}>
        <ErrorIcon />
        {label}
    </div>
)
export const ChangesetStatusProcessing: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Processing</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <TimerSandIcon className="changeset-status-cell__processing-icon" />
        {label}
    </div>
)
