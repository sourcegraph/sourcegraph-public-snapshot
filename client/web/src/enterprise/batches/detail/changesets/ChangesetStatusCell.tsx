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
    mdiDotsVertical,
} from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { Tooltip, Icon } from '@sourcegraph/wildcard'

import { type ChangesetFields, ChangesetState, type Scalars } from '../../../../graphql-operations'

import { ChangesetStatusScheduled } from './ChangesetStatusScheduled'

export interface ChangesetStatusCellProps {
    className?: string
    id?: Scalars['ID']
    state: ChangesetFields['state']
    role?: React.AriaRole
}

export const ChangesetStatusCell: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusCellProps>> = ({
    id,
    state,
    className = 'd-flex',
    role,
}) => {
    switch (state) {
        case ChangesetState.FAILED: {
            return <ChangesetStatusError className={className} role={role} />
        }
        case ChangesetState.RETRYING: {
            return <ChangesetStatusRetrying className={className} role={role} />
        }
        case ChangesetState.SCHEDULED: {
            return <ChangesetStatusScheduled className={className} role={role} id={id} />
        }
        case ChangesetState.PROCESSING: {
            return <ChangesetStatusProcessing className={className} role={role} />
        }
        case ChangesetState.UNPUBLISHED: {
            return <ChangesetStatusUnpublished className={className} role={role} />
        }
        case ChangesetState.OPEN: {
            return <ChangesetStatusOpen className={className} role={role} />
        }
        case ChangesetState.DRAFT: {
            return <ChangesetStatusDraft className={className} role={role} />
        }
        case ChangesetState.CLOSED: {
            return <ChangesetStatusClosed className={className} role={role} />
        }
        case ChangesetState.MERGED: {
            return <ChangesetStatusMerged className={className} role={role} />
        }
        case ChangesetState.READONLY: {
            return <ChangesetStatusReadOnly className={className} role={role} />
        }
        case ChangesetState.DELETED: {
            return <ChangesetStatusDeleted className={className} role={role} />
        }
    }
}

const iconClassNames = 'm-0 text-nowrap flex-column align-items-center justify-content-center'

const StatusLabel: React.FunctionComponent<{ status: string; className?: string }> = ({ status, className }) => (
    // Relative positioning needed to avoid VisuallyHidden creating a double layer scrollbar in Chrome.
    // Related bug: https://bugs.chromium.org/p/chromium/issues/detail?id=1154640#c15
    <span className={classNames(className, 'position-relative')}>
        <VisuallyHidden>Status:</VisuallyHidden> {status}
    </span>
)

interface ChangesetStatusIconProps extends React.HTMLAttributes<HTMLDivElement> {
    label?: React.ReactNode
    className?: string
}

export const ChangesetStatusUnpublished: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Unpublished" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiSourceBranch} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusClosed: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Closed" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-danger" svgPath={mdiSourcePull} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusMerged: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Merged" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-merged" svgPath={mdiSourceMerge} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusOpen: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Open" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-success" svgPath={mdiSourcePull} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusDraft: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Draft" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiSourcePull} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusDeleted: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Deleted" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiDelete} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusError: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel className="text-danger" status="Failed" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon className="text-danger" svgPath={mdiAlertCircle} inline={false} aria-hidden={true} />
        {label}
    </div>
)
export const ChangesetStatusRetrying: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Retrying" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiAutorenew} inline={false} aria-hidden={true} />
        {label}
    </div>
)

export const ChangesetStatusProcessing: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Processing" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiTimerSand} inline={false} aria-hidden={true} />
        {label}
    </div>
)

export const ChangesetStatusArchived: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Archived" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiArchive} inline={false} aria-hidden={true} />
        {label}
    </div>
)

export const ChangesetStatusReadOnly: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Read-only" />,
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

export const ChangesetStatusOthers: React.FunctionComponent<React.PropsWithChildren<ChangesetStatusIconProps>> = ({
    label = <StatusLabel status="Others" />,
    className,
    ...props
}) => (
    <div className={classNames(iconClassNames, className)} {...props}>
        <Icon svgPath={mdiDotsVertical} inline={false} aria-hidden={true} />
        {label}
    </div>
)
