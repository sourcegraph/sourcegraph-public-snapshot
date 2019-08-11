import React from 'react'
import { Link } from 'react-router-dom'
import { LinkOrSpan } from '../../../../../shared/src/components/LinkOrSpan'
import { displayRepoName } from '../../../../../shared/src/components/RepoFileLink'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadStateIcon } from '../common/threadState/ThreadStateIcon'

interface Props {
    thread: Pick<GQL.ThreadOrThreadPreview, '__typename' | 'title' | 'kind' | 'repository'> &
        (Pick<GQL.IThread, 'number' | 'state' | 'url'> | { number?: undefined; state?: undefined; url?: undefined })

    showRepository?: boolean
}

/**
 * An item in the list of threads.
 */
export const ThreadListItem: React.FunctionComponent<Props> = ({ thread, showRepository }) => (
    <LinkOrSpan to={thread.url} className="d-flex align-items-center text-decoration-none">
        <ThreadStateIcon
            thread={thread.state ? thread : { kind: thread.kind, state: GQL.ThreadState.OPEN }}
            className="mr-2"
        />
        <span className="text-truncate">
            {thread.number !== undefined ? (
                <span className="text-muted mr-2">
                    {showRepository && (
                        <>
                            <Link to={thread.repository.url}>{displayRepoName(thread.repository.name)}</Link>/
                        </>
                    )}
                    #{thread.number}
                </span>
            ) : (
                <span className="text-muted mr-2">
                    New {thread.kind.toLowerCase()} in{' '}
                    <Link to={thread.repository.url}>{displayRepoName(thread.repository.name)}</Link>:
                </span>
            )}
            {thread.title}
        </span>
    </LinkOrSpan>
)
