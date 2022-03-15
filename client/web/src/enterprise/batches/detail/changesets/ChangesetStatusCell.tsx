import React from 'react'

import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import AutorenewIcon from 'mdi-react/AutorenewIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'

import { ChangesetFields, ChangesetState, Scalars } from '../../../../graphql-operations'

import { ChangesetStatusScheduled } from './ChangesetStatusScheduled'

export interface ChangesetStatusCellProps {
    className?: string
    id?: Scalars['ID']
    state: ChangesetFields['state']
}

export const ChangesetStatusCell: React.FunctionComponent<ChangesetStatusCellProps> = ({
    id,
    state,
    className = 'd-flex',
}) => {
    switch (state) {
        case ChangesetState.FAILED:
            return <ChangesetStatusError className={className} />
        case ChangesetState.RETRYING:
            return <ChangesetStatusRetrying className={className} />
        case ChangesetState.SCHEDULED:
            return <ChangesetStatusScheduled className={className} id={id} />
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

interface ChangesetStatusIconProps {
    label?: React.ReactNode
    className?: string
}

export const ChangesetStatusUnpublished: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Unpublished</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <SourceBranchIcon />
        {label}
    </div>
)
export const ChangesetStatusClosed: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Closed</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <SourcePullIcon className="text-danger" />
        {label}
    </div>
)
export const ChangesetStatusMerged: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Merged</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <SourceMergeIcon className="text-merged" />
        {label}
    </div>
)
export const ChangesetStatusOpen: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Open</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <SourcePullIcon className="text-success" />
        {label}
    </div>
)
export const ChangesetStatusDraft: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Draft</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <SourcePullIcon />
        {label}
    </div>
)
export const ChangesetStatusDeleted: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Deleted</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <DeleteIcon />
        {label}
    </div>
)
export const ChangesetStatusError: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span className="text-danger">Failed</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <AlertCircleIcon className="text-danger" />
        {label}
    </div>
)
export const ChangesetStatusRetrying: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Retrying</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <AutorenewIcon />
        {label}
    </div>
)

export const ChangesetStatusProcessing: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Processing</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <TimerSandIcon />
        {label}
    </div>
)

export const ChangesetStatusArchived: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Archived</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <ArchiveIcon />
        {label}
    </div>
)
