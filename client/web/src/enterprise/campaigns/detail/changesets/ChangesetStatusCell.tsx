import React from 'react'
import { ChangesetFields, ChangesetState } from '../../../../graphql-operations'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import AutorenewIcon from 'mdi-react/AutorenewIcon'
import ErrorIcon from 'mdi-react/ErrorIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import classNames from 'classnames'

export interface ChangesetStatusCellProps {
    className?: string
    changeset: Pick<ChangesetFields, 'state'>
}

export const ChangesetStatusCell: React.FunctionComponent<ChangesetStatusCellProps> = ({
    changeset,
    className = 'd-flex',
}) => {
    switch (changeset.state) {
        case ChangesetState.FAILED:
            return <ChangesetStatusError className={className} />
        case ChangesetState.RETRYING:
            return <ChangesetStatusRetrying className={className} />
        case ChangesetState.PROCESSING:
            return <ChangesetStatusProcessing className={className} />
        case ChangesetState.UNPUBLISHED:
            return <ChangesetStatusUnpublished className={className} />
        case ChangesetState.OPEN:
            return <ChangesetStatusOpen className={className} />
        case ChangesetState.DRAFT:
            return <ChangesetStatusDraft className={className} />
        case ChangesetState.CLOSED:
            return <ChangesetStatusClosed className={className} />
        case ChangesetState.MERGED:
            return <ChangesetStatusMerged className={className} />
        case ChangesetState.DELETED:
            return <ChangesetStatusDeleted className={className} />
    }
}

const iconClassNames = 'm-0 text-nowrap flex-column align-items-center justify-content-center'

export const ChangesetStatusUnpublished: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Unpublished</span>,
    className,
}) => (
    <div
        className={classNames(iconClassNames, 'text-muted', className)}
        data-tooltip="Set published: true to publish to code host"
    >
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
export const ChangesetStatusDraft: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Draft</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-muted', className)}>
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
    label = <span className="text-danger">Failed</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-danger', className)}>
        <ErrorIcon />
        {label}
    </div>
)
export const ChangesetStatusRetrying: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Retrying</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, 'text-muted', className)}>
        <AutorenewIcon />
        {label}
    </div>
)
export const ChangesetStatusProcessing: React.FunctionComponent<{ label?: JSX.Element; className?: string }> = ({
    label = <span className="text-muted">Processing</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <TimerSandIcon />
        {label}
    </div>
)
