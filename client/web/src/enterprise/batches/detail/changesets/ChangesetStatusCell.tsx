import React from 'react'

import {
    mdiSourceBranch,
    mdiSourcePull,
    mdiSourceMerge,
    mdiDelete,
    mdiAlertCircle,
    mdiAutorenew,
    mdiTimerSand,
    mdiArchive,
    mdiLock,
} from '@mdi/js'
import classNames from 'classnames'

import { Tooltip, Icon } from '@sourcegraph/wildcard'

import { ChangesetFields, ChangesetState, Scalars } from '../../../../graphql-operations'

import { ChangesetStatusScheduled } from './ChangesetStatusScheduled'

export interface ChangesetStatusCellProps {
    className?: string
    id?: Scalars['ID']
    state: ChangesetFields['state']
}

export const ChangesetStatusCell: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusCellProps>> = ({
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
        case ChangesetState.READONLY:
            return <ChangesetStatusReadOnly className={className} />
        case ChangesetState.DELETED:
            return <ChangesetStatusDeleted className={className} />
    }
}

const iconClassNames = 'm-0 text-nowrap flex-column align-items-center justify-content-center'

interface ChangesetStatusIconProps extends React.HTMLAttributes<HTMLDivElement> {
    label?: React.ReactNode
    className?: string
}

export const ChangesetStatusUnpublished: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Unpublished</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiSourceBranch} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusClosed: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Closed</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-danger" svgPath={mdiSourcePull} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusMerged: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Merged</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-merged" svgPath={mdiSourceMerge} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusOpen: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Open</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-success" svgPath={mdiSourcePull} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusDraft: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Draft</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiSourcePull} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusDeleted: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Deleted</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiDelete} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusError: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span className="text-danger">Failed</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-danger" svgPath={mdiAlertCircle} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusRetrying: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Retrying</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiAutorenew} inline={false} aria-hidden={true} />
        {label}
    </div>
)

export const ChangesetStatusProcessing: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Processing</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiTimerSand} inline={false} aria-hidden={true} />
        {label}
    </div>
)

export const ChangesetStatusArchived: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Archived</span>,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiArchive} inline={false} aria-hidden={true} />
        {label}
    </div>
)

export const ChangesetStatusReadOnly: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <span>Read-only</span>,
    className,
    ...props
}) => (
    <Tooltip content="This changeset is read-only, and cannot be modified. This is usually caused by the repository being archived.">
        <div className={classNames(iconClassNames, className)} {...props}>
            <Icon svgPath={mdiLock} inline={false} aria-hidden={true} />
            {label}
        </div>
    </Tooltip>
)
