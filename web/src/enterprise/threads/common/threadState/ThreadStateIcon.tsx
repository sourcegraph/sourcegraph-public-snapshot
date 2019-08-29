import React from 'react'
import { ThreadStateFields, threadStateInfo } from './threadState'

interface Props {
    thread: ThreadStateFields
    className?: string
}

/**
 * An icon that indicates the status of a thread.
 */
export const ThreadStateIcon: React.FunctionComponent<Props> = ({ thread, className = '' }) => {
    const { color, icon: Icon, text } = threadStateInfo(thread)
    return <Icon className={`icon-inline text-${color} ${className}`} data-tooltip={text} />
}
