import React from 'react'
import { ThreadStatusFields, threadStatusInfo } from './threadStatus'

interface Props {
    thread: ThreadStatusFields
    className?: string
}

/**
 * An icon that indicates the status of a thread.
 */
export const ThreadStatusIcon: React.FunctionComponent<Props> = ({ thread, className = '' }) => {
    const { color, icon: Icon, text } = threadStatusInfo(thread)
    return <Icon className={`icon-inline text-${color} ${className}`} data-tooltip={text} />
}
