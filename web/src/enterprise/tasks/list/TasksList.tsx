import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { WithTasksQueryResults } from '../components/withTasksQueryResults/WithTasksQueryResults'
import { TasksAreaContext } from '../global/TasksArea'
import { TasksListItem } from './item/TasksListItem'

export interface TasksListContext {
    containerClassName?: string
}

interface Props
    extends QueryParameterProps,
        TasksListContext,
        TasksAreaContext,
        ExtensionsControllerProps,
        PlatformContextProps {
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * The list of tasks with a header.
 */
export const TasksList: React.FunctionComponent<Props> = ({ containerClassName, query, ...props }) => (
    <WithTasksQueryResults {...props} query={query}>
        {({ tasksOrError }) => (
            <div className="tasks-list">
                {isErrorLike(tasksOrError) ? (
                    <div className={containerClassName}>
                        <div className="alert alert-danger mt-2">{tasksOrError.message}</div>
                    </div>
                ) : tasksOrError === LOADING ? (
                    <div className={containerClassName}>
                        <LoadingSpinner className="mt-3" />
                    </div>
                ) : tasksOrError.length === 0 ? (
                    <div className={containerClassName}>
                        <p className="p-2 mb-0 text-muted">No tasks found.</p>
                    </div>
                ) : (
                    <ul className="list-group list-group-flush mb-0">
                        {tasksOrError.map((task, i) => (
                            <li key={i} className="list-group-item px-0">
                                <TasksListItem {...props} key={i} diagnostic={task} className={containerClassName} />
                            </li>
                        ))}
                    </ul>
                )}
                <style>
                    {/* HACK TODO!(sqs) */}
                    {
                        '.tasks-list .markdown pre,.tasks-list .markdown code {margin:0;padding:0;background-color:transparent;}'
                    }
                </style>
            </div>
        )}
    </WithTasksQueryResults>
)
