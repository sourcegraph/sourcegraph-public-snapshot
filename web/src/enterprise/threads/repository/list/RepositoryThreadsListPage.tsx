import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { useQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThreadListFilterContext } from '../../list/header/ThreadListFilterDropdownButton'
import { ThreadList, ThreadListContext, ThreadListHeaderCommonFilters } from '../../list/ThreadList'
import { useThreads } from '../../list/useThreads'
import { RepositoryThreadsAreaContext } from '../RepositoryThreadsArea'

interface Props
    extends Pick<RepositoryThreadsAreaContext, 'repo'>,
        ThreadListContext,
        ExtensionsControllerNotificationProps {
    newThreadURL: string | null
}

/**
 * Lists a repository's threads.
 */
export const RepositoryThreadsListPage: React.FunctionComponent<Props> = ({ newThreadURL, repo, ...props }) => {
    const [query, onQueryChange] = useQueryParameter(props)
    const threads = useThreads({ query, repositories: [repo.id] })
    const filterProps: ThreadListFilterContext = {
        threadConnection: threads,
        query,
        onQueryChange,
    }
    return (
        <div className="repository-threads-list-page">
            {newThreadURL && (
                <Link to={newThreadURL} className="btn btn-primary mb-3">
                    New thread
                </Link>
            )}
            <ThreadList
                {...props}
                threads={threads}
                query={query}
                onQueryChange={onQueryChange}
                itemCheckboxes={true}
                headerItems={{
                    right: (
                        <>
                            <ThreadListHeaderCommonFilters {...filterProps} />
                        </>
                    ),
                }}
            />
        </div>
    )
}
