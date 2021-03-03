import React from 'react'
import classnames from 'classnames'

export type BadgeStatus = 'beta' | 'prototype' | 'wip' | 'new'

const statusStyleMapping: Record<BadgeStatus, string> = {
    prototype: 'badge-warning',
    wip: 'badge-warning',
    beta: 'badge-info',
    new: 'badge-info',
}

interface Props {
    status: BadgeStatus
    /** Render a mailto href to share feedback */
    feedback?: {
        mailto: string
        /** Defaults to 'Share feedback' */
        text?: string
    }
    tooltip?: string
    className?: string
}

export const StatusBadge: React.FC<Props> = props => {
    const { className, status, feedback, tooltip } = props

    return (
        <div className={classnames('d-flex', 'align-items-center', className)}>
            <span className={classnames('badge', 'text-uppercase', statusStyleMapping[status])} data-tooltip={tooltip}>
                {status}
            </span>
            {feedback && (
                <a href={`mailto:${feedback.mailto}`} className="ml-2" target="_blank" rel="noopener noreferrer">
                    {feedback.text || 'Share feedback'}
                </a>
            )}
        </div>
    )
}
