import React from 'react'
import { changeStatusInfo, ChangeWithStatus } from './changeStatus'

interface Props {
    change: ChangeWithStatus
    className?: string
}

/**
 * A badge that displays the status of a change.
 */
export const ChangeStatusBadge: React.FunctionComponent<Props> = ({ change, className = '' }) => {
    const { color, icon: Icon, text } = changeStatusInfo(change)
    return (
        <span
            className={`badge badge-${color} ${className} d-inline-flex align-items-center py-1 px-2 h6 mb-0 font-weight-bold`}
        >
            <Icon className="icon-inline mr-1" /> {text}
        </span>
    )
}
