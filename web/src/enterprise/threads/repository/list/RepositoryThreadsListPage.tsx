import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ConnectionListFilterContext } from '../../../../components/connectionList/ConnectionListFilterDropdownButton'
import { useQueryParameter } from '../../../../util/useQueryParameter'
import { ThreadList, ThreadListContext, ThreadListHeaderCommonFilters } from '../../list/ThreadList'
import { useThreads } from '../../list/useThreads'
import { RepositoryThreadsAreaContext } from '../RepositoryThreadsArea'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { ThemeProps } from '../../../../theme'

interface Props
    extends Pick<RepositoryThreadsAreaContext, 'repo'>,
        ThreadListContext,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    newThreadURL: string | null
}

/**
 * Lists a repository's threads.
 */
export const RepositoryThreadsListPage: React.FunctionComponent<Props> = ({ newThreadURL, repo, ...props }) => {
    const [query, onQueryChange, locationWithQuery] = useQueryParameter(props)
    const threads = useThreads({ query, repositories: [repo.id] })
    const filterProps: ConnectionListFilterContext<GQL.IThreadConnectionFilters> = {
        connection: threads,
        query,
        onQueryChange,
        locationWithQuery,
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
                locationWithQuery={locationWithQuery}
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
