import React from 'react'
import { ThreadStateFields, threadStateInfo } from './threadState'

interface Props {
    thread: ThreadStateFields
    className?: string
}

/**
 * A badge that displays the state of a thread.
 */
export const ThreadStateBadge: React.FunctionComponent<Props> = ({ thread, className = '' }) => {
    const { color, icon: Icon, text } = threadStateInfo(thread)
    return (
        <span
            className={`badge badge-${color} ${className} d-inline-flex align-items-center py-2 px-4 h6 mb-0 font-weight-bold`}
        >
            <Icon className="icon-inline mr-1" /> {text}
        </span>
    )
}
