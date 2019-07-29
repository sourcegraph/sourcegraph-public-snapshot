import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../../shared/src/util/strings'
import { ThreadAreaContext } from '../ThreadArea'
import { AddThreadToThreadDropdownButton } from './AddThreadToThreadDropdownButton'
import { ThreadThreadListItem } from './ThreadThreadListItem'
import { useThreadThreads } from './useThreadThreads'

interface Props extends Pick<ThreadAreaContext, 'thread'>, ExtensionsControllerProps {
    className?: string
}

const LOADING = 'loading' as const

export const ThreadThreadsListPage: React.FunctionComponent<Props> = ({ thread, className = '', ...props }) => {
    const [threadsOrError, onThreadsUpdate] = useThreadThreads(thread)

    return (
        <div className={`thread-threads-list-page ${className}`}>
            <AddThreadToThreadDropdownButton
                {...props}
                thread={thread}
                onAdd={onThreadsUpdate}
                className="mb-3"
            />
            {threadsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(threadsOrError) ? (
                <div className="alert alert-danger">{threadsOrError.message}</div>
            ) : (
                <div className="card">
                    <div className="card-header">
                        <span className="text-muted">
                            {threadsOrError.totalCount} {pluralize('thread', threadsOrError.totalCount)}
                        </span>
                    </div>
                    {threadsOrError.nodes.length > 0 ? (
                        <ul className="list-group list-group-flush">
                            {threadsOrError.nodes.map(thread => (
                                <li key={thread.id} className="list-group-item">
                                    <ThreadThreadListItem
                                        {...props}
                                        thread={thread}
                                        thread={thread}
                                        onUpdate={onThreadsUpdate}
                                    />
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <div className="p-2 text-muted">No threads.</div>
                    )}
                </div>
            )}
        </div>
    )
}
