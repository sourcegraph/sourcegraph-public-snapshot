import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { ThreadsList } from '../../list/ThreadsList'
import { useThreads } from '../../list/useThreads'
import { RepositoryThreadsAreaContext } from '../RepositoryThreadsArea'

interface Props extends Pick<RepositoryThreadsAreaContext, 'repo'>, ExtensionsControllerNotificationProps {
    newThreadURL: string | null
}

/**
 * Lists a repository's threads.
 */
export const RepositoryThreadsListPage: React.FunctionComponent<Props> = ({ newThreadURL, repo, ...props }) => {
    const threads = useThreads(repo)
    return (
        <div className="repository-threads-list-page">
            {newThreadURL && (
                <Link to={newThreadURL} className="btn btn-primary mb-3">
                    New thread
                </Link>
            )}
            <ThreadsList {...props} threads={threads} />
        </div>
    )
}
