import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { ThreadListItem } from './ThreadListItem'

const LOADING: 'loading' = 'loading'

interface Props extends ExtensionsControllerNotificationProps {
    threads: typeof LOADING | GQL.IThreadConnection | ErrorLike
}

/**
 * Lists threads.
 */
export const ThreadsList: React.FunctionComponent<Props> = ({ threads, ...props }) => (
    <div className="threads-list">
        {threads === LOADING ? (
            <LoadingSpinner className="icon-inline mt-3" />
        ) : isErrorLike(threads) ? (
            <div className="alert alert-danger mt-3">{threads.message}</div>
        ) : (
            <div className="card">
                <div className="card-header">
                    <span className="text-muted">
                        {threads.totalCount} {pluralize('thread', threads.totalCount)}
                    </span>
                </div>
                {threads.nodes.length > 0 ? (
                    <ul className="list-group list-group-flush">
                        {threads.nodes.map(thread => (
                            <li key={thread.id} className="list-group-item">
                                <ThreadListItem {...props} thread={thread} />
                            </li>
                        ))}
                    </ul>
                ) : (
                    <div className="p-2 text-muted">No threads yet.</div>
                )}
            </div>
        )}
    </div>
)
