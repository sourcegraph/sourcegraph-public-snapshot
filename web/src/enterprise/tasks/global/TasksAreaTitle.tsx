import React from 'react'
import { TasksIcon } from '../icons'

interface Props {
    className?: string
    primaryActions?: JSX.Element | null
}

/**
 * The tasks area title.
 *
 * // TODO!(sqs): dedupe with ChecksAreaTitle?
 */
export const TasksAreaTitle: React.FunctionComponent<Props> = ({ className = '', primaryActions, children }) => (
    <div className="d-flex align-items-center mb-3">
        <h1 className={`h3 mb-0 d-flex align-items-center ${className}`}>
            <TasksIcon className="icon-inline mr-1" /> Tasks
        </h1>
        {children}
        {primaryActions && (
            <>
                <div className="flex-1" />
                {primaryActions}
            </>
        )}
    </div>
)
