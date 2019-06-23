import UnfoldLessVerticalIcon from 'mdi-react/UnfoldLessVerticalIcon'
import UnfoldMoreVerticalIcon from 'mdi-react/UnfoldMoreVerticalIcon'
import React, { useCallback, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { PageTitle } from '../../../components/PageTitle'
import { WithQueryParameter } from '../../../components/withQueryParameter/WithQueryParameter'
import { TasksAreaContext } from '../global/TasksArea'
import { TasksAreaTitle } from '../components/TasksAreaTitle'
import { TasksList } from './TasksList'

interface Props extends TasksAreaContext, RouteComponentProps<{}> {}

const LOCAL_STORAGE_KEY = 'ChecksDashboardPage-container'

/**
 * The tasks list page.
 */
export const TasksListPage: React.FunctionComponent<Props> = ({ match, ...props }) => {
    const initialIsExpanded = useMemo(() => localStorage.getItem(LOCAL_STORAGE_KEY) !== null, [])
    const [isExpanded, setIsExpanded] = useState(initialIsExpanded)
    const toggleIsExpanded = useCallback(() => {
        setIsExpanded(!isExpanded)
        if (isExpanded) {
            localStorage.removeItem(LOCAL_STORAGE_KEY)
        } else {
            localStorage.setItem(LOCAL_STORAGE_KEY, 'expanded')
        }
    }, [isExpanded])

    const containerClassName = isExpanded ? 'container-fluid' : 'container'

    return (
        <div className="w-100">
            <PageTitle title="Tasks" />
            <div className={`${containerClassName} mt-3`}>
                <TasksAreaTitle>
                    <button
                        type="button"
                        className="btn btn-link text-decoration-none"
                        data-tooltip={isExpanded ? 'Exit widescreen view' : 'Enter widescreen view'}
                        onClick={toggleIsExpanded}
                    >
                        {isExpanded ? (
                            <UnfoldLessVerticalIcon className="icon-inline" />
                        ) : (
                            <UnfoldMoreVerticalIcon className="icon-inline" />
                        )}
                    </button>
                </TasksAreaTitle>
            </div>
            <WithQueryParameter {...props}>
                {({ query, onQueryChange }) => (
                    <TasksList
                        {...props}
                        query={query}
                        onQueryChange={onQueryChange}
                        containerClassName={containerClassName}
                    />
                )}
            </WithQueryParameter>
        </div>
    )
}
