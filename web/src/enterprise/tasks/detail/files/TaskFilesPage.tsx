import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { WithQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { Task } from '../../task'
import { TaskFilesList } from './TaskFilesList'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    task: Task

    history: H.History
    location: H.Location
    isLightTheme: boolean
}

/**
 * The "Files changed" page for a task.
 */
export const TaskFilesPage: React.FunctionComponent<Props> = ({ task, ...props }) => (
    <div className="task-files-page">
        <WithQueryParameter defaultQuery={/* TODO!(sqs) */ ''} history={props.history} location={props.location}>
            {({ query, onQueryChange }) => (
                <TaskFilesList {...props} task={task} query={query} onQueryChange={onQueryChange} />
            )}
        </WithQueryParameter>
    </div>
)
