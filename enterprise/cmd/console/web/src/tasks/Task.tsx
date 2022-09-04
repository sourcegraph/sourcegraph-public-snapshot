import React from 'react'

import { Task as TaskData } from './taskProvider'

export const Task: React.FunctionComponent<{
    task: TaskData
    tag?: 'li'
}> = ({ task, tag: Tag = 'li' }) => {
    const subtasks = ['x', 'y', 'z']
    return (
        <Tag className="task-group">
            <div className="container">
                <h2>
                    <a href={task.url}>{task.text}</a>
                </h2>
                <ol className="tasks">
                    {subtasks.map(task => (
                        <li key={task} className="task">
                            <span className="date">Due|Assigned</span>
                            <summary>{task}</summary>
                        </li>
                    ))}
                </ol>
            </div>
        </Tag>
    )
}
