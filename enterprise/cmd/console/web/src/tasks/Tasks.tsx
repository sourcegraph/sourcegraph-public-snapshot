import './Tasks.css'

import { useObservable, useObservableGetState, useObservableState } from 'observable-hooks'
import React, { useMemo } from 'react'
import { combineLatest, flatMap, from, map, NEVER, of } from 'rxjs'

import { SettingsProps } from '../app/useSettings'
import { useGoogleAPIScript } from '../services/google/useGoogleAPI'
import { Task } from './Task'
import { TaskProvider } from './taskProvider'

export const Tasks: React.FunctionComponent<
    {
        taskProviders: TaskProvider[]
        tag?: 'main'
        className?: string
    } & Pick<SettingsProps, 'settings'>
> = ({ taskProviders, settings, tag: Tag = 'main', className }) => {
    // TODO(sqs): init this another way that is only in the google providers
    const googleAPIReady = useGoogleAPIScript()

    const tasks = useObservableState(
        useMemo(
            () =>
                googleAPIReady
                    ? combineLatest(taskProviders.map(({ name, tasks }) => from(tasks(settings)))).pipe(
                          map(tasks => tasks.flat())
                      )
                    : NEVER,
            [googleAPIReady, settings, taskProviders]
        )
    )

    return (
        <Tag className={className}>
            {tasks === undefined ? (
                <div className="container">
                    <p>Loading...</p>
                </div>
            ) : tasks.length === 0 ? (
                <div className="container">
                    <p>No tasks found.</p>
                </div>
            ) : (
                <ol className="task-groups">
                    {tasks.map((task, i) => (
                        <Task key={i} task={task} tag="li" />
                    ))}
                </ol>
            )}
        </Tag>
    )
}
