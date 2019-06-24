import React from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { WithQueryParameter } from '../../../components/withQueryParameter/WithQueryParameter'
import { TasksAreaTitle } from '../components/TasksAreaTitle'
import { TasksAreaContext } from '../global/TasksArea'
import { TasksList } from './TasksList'

interface Props extends TasksAreaContext, RouteComponentProps<{}> {}

/**
 * The tasks list page.
 */
export const TasksListPage: React.FunctionComponent<Props> = ({ match, ...props }) => (
    <div className="w-100 mt-3">
        <PageTitle title="Tasks" />
        <div className="container-fluid">
            <TasksAreaTitle />
        </div>
        <WithQueryParameter {...props}>
            {({ query, onQueryChange }) => (
                <TasksList
                    {...props}
                    query={query}
                    onQueryChange={onQueryChange}
                    containerClassName="container-fluid"
                />
            )}
        </WithQueryParameter>
    </div>
)
