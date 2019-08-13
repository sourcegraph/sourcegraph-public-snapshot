import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../components/withQueryParameter/WithQueryParameter'
import { ThreadListHeader, ThreadListHeaderItems } from './ThreadListHeader'
import { ThreadListItem, ThreadListItemContext } from './ThreadListItem'

export interface ThreadListContext extends ThreadListItemContext {
    /**
     * Whether each item should have a checkbox.
     */
    itemCheckboxes?: boolean

    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

interface Props extends QueryParameterProps, ThreadListContext {
    threads: typeof LOADING | GQL.IThreadConnection | GQL.IThreadOrThreadPreviewConnection | ErrorLike

    headerItems?: ThreadListHeaderItems

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The list of threads with a header.
 */
export const ThreadList: React.FunctionComponent<Props> = ({
    threads,
    headerItems,
    itemCheckboxes,
    query,
    onQueryChange,
    className = '',
    ...props
}) => (
    <div className={`thread-list ${className}`}>
        {isErrorLike(threads) ? (
            <div className="alert alert-danger mt-2">{threads.message}</div>
        ) : (
            <div className="card">
                <ThreadListHeader
                    {...props}
                    threads={threads}
                    items={headerItems}
                    itemCheckboxes={itemCheckboxes}
                    query={query}
                />
                {threads === LOADING ? (
                    <LoadingSpinner className="m-3" />
                ) : threads.nodes.length === 0 ? (
                    <p className="p-2 mb-0 text-muted">No threads found.</p>
                ) : (
                    <ul className="list-group list-group-flush">
                        {(threads.nodes as GQL.ThreadOrThreadPreview[]).map((thread, i) => (
                            <ThreadListItem key={i} {...props} thread={thread} itemCheckboxes={itemCheckboxes} />
                        ))}
                    </ul>
                )}
            </div>
        )}
    </div>
)
