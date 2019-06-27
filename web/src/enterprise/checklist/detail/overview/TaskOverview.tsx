import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { Task } from '../../task'
import { TaskDescription } from './TaskDescription'

interface Props extends ExtensionsControllerProps {
    task: Task

    areaURL: string

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The overview for a single task.
 */
export const TaskOverview: React.FunctionComponent<Props> = ({ task, areaURL, className = '', ...props }) => (
    <div className={`task-overview ${className || ''}`}>
        <TaskDescription {...props} task={task} />
    </div>
)
