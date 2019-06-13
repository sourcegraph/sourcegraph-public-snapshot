import React from 'react'
import { changeStatusInfo, ChangeWithStatus } from './changeStatus'

interface Props {
    change: ChangeWithStatus
    className?: string
}

/**
 * An icon that indicates the status of a change.
 */
export const ChangeStatusIcon: React.FunctionComponent<Props> = ({ change, className = '' }) => {
    const { color, icon: Icon, text } = changeStatusInfo(change)
    return <Icon className={`icon-inline text-${color} ${className}`} data-tooltip={text} />
}
