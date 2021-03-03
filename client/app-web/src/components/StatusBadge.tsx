import React from 'react'

type Status = 'beta' | 'prototype' | 'wip'

const statusStyleMapping: Record<Status, string> = {
    prototype: 'badge-warning',
    wip: 'badge-warning',
    beta: 'badge-info',
}

interface Props {
    status: Status
    /** Render a mailto href to share feedback */
    feedback?: {
        mailto: string
        /** Defaults to 'Share feedback' */
        text?: string
    }
    tooltip?: string
}

export const StatusBadge: React.FunctionComponent<Props> = ({ status, feedback, tooltip }) => (
    <div className="d-flex align-items-center">
        <span className={`badge ${statusStyleMapping[status]} text-uppercase mr-2`} data-tooltip={tooltip}>
            {status}
        </span>
        {feedback && (
            <a href={`mailto:${feedback.mailto}`} target="_blank" rel="noopener noreferrer">
                {feedback.text || 'Share feedback'}
            </a>
        )}
    </div>
)
