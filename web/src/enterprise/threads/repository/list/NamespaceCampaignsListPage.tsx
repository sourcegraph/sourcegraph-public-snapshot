import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { NamespaceAreaContext } from '../../../../namespaces/NamespaceArea'
import { ThreadsList } from '../../list/ThreadsList'
import { useThreads } from '../../list/useThreads'

interface Props extends Pick<NamespaceAreaContext, 'namespace'>, ExtensionsControllerNotificationProps {
    newThreadURL: string | null
}

/**
 * Lists a namespace's threads.
 */
export const NamespaceThreadsListPage: React.FunctionComponent<Props> = ({ newThreadURL, namespace, ...props }) => {
    const threads = useThreads(namespace)
    return (
        <div className="namespace-threads-list-page">
            {newThreadURL && (
                <Link to={newThreadURL} className="btn btn-primary mb-3">
                    New thread
                </Link>
            )}
            <ThreadsList {...props} threads={threads} />
        </div>
    )
}
