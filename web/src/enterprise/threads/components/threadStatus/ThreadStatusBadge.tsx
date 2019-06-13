import React from 'react'
import { ThreadStatusFields, threadStatusInfo } from './threadStatus'

interface Props {
    thread: ThreadStatusFields
    className?: string
}

/**
 * A badge that displays the status of a thread.
 */
export const ThreadStatusBadge: React.FunctionComponent<Props> = ({ thread, className = '' }) => {
    const { color, icon: Icon, text } = threadStatusInfo(thread)
    return (
        <span
            className={`badge badge-${color} ${className} d-inline-flex align-items-center py-2 px-4 h6 mb-0 font-weight-bold`}
        >
            <Icon className="icon-inline mr-1" /> {text}
        </span>
    )
}
