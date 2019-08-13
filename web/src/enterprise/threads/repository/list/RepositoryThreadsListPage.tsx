import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ConnectionListFilterContext } from '../../../../components/connectionList/ConnectionListFilterDropdownButton'
import { useQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
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
    const filterProps: ConnectionListFilterContext<GQL.IThreadConnectionFilters> = {
        connection: threads,
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
