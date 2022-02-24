import classNames from 'classnames'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import ArchiveIcon from 'mdi-react/ArchiveIcon'
import AutorenewIcon from 'mdi-react/AutorenewIcon'
import DeleteIcon from 'mdi-react/DeleteIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import SourceMergeIcon from 'mdi-react/SourceMergeIcon'
import SourcePullIcon from 'mdi-react/SourcePullIcon'
import TimerSandIcon from 'mdi-react/TimerSandIcon'
import React from 'react'

import { Icon } from '@sourcegraph/wildcard'

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
            return <Icon inline={false} as={ChangesetStatusError} className={className} />
        case ChangesetState.RETRYING:
            return <Icon inline={false} as={ChangesetStatusRetrying} className={className} />
        case ChangesetState.SCHEDULED:
            return <Icon inline={false} as={ChangesetStatusScheduled} className={className} />
        case ChangesetState.PROCESSING:
            return <Icon inline={false} as={ChangesetStatusProcessing} className={className} />
        case ChangesetState.UNPUBLISHED:
            return <Icon inline={false} as={ChangesetStatusUnpublished} className={className} />
        case ChangesetState.OPEN:
            return <Icon inline={false} as={ChangesetStatusOpen} className={className} />
        case ChangesetState.DRAFT:
            return <Icon inline={false} as={ChangesetStatusDraft} className={className} />
        case ChangesetState.CLOSED:
            return <Icon inline={false} as={ChangesetStatusClosed} className={className} />
        case ChangesetState.MERGED:
            return <Icon inline={false} as={ChangesetStatusMerged} className={className} />
        case ChangesetState.DELETED:
            return <Icon inline={false} as={ChangesetStatusDeleted} className={className} />
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
        <Icon inline={false} as={SourceBranchIcon} />
        {label}
    </div>
)
export const ChangesetStatusClosed: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Closed</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={SourcePullIcon} />
        {label}
    </div>
)
export const ChangesetStatusMerged: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Merged</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={SourceMergeIcon} />
        {label}
    </div>
)
export const ChangesetStatusOpen: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Open</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={SourcePullIcon} />
        {label}
    </div>
)
export const ChangesetStatusDraft: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Draft</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={SourcePullIcon} />
        {label}
    </div>
)
export const ChangesetStatusDeleted: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Deleted</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={DeleteIcon} />
        {label}
    </div>
)
export const ChangesetStatusError: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span className="text-danger">Failed</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={AlertCircleIcon} />
        {label}
    </div>
)
export const ChangesetStatusRetrying: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Retrying</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={AutorenewIcon} />
        {label}
    </div>
)

export const ChangesetStatusProcessing: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Processing</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={TimerSandIcon} />
        {label}
    </div>
)

export const ChangesetStatusArchived: React.FunctionComponent<ChangesetStatusIconProps> = ({
    label = <span>Archived</span>,
    className,
}) => (
    <div className={classNames(iconClassNames, className)}>
        <Icon inline={false} as={ArchiveIcon} />
        {label}
    </div>
)
