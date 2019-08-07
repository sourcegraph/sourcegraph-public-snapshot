import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadStateIcon } from '../common/threadState/ThreadStateIcon'

interface Props {
    thread: Pick<GQL.IThread, '__typename' | 'number' | 'title' | 'state' | 'url'>
}

/**
 * An item in the list of threads.
 */
export const ThreadListItem: React.FunctionComponent<Props> = ({ thread }) => (
    <Link to={thread.url} className="d-flex align-items-center text-decoration-none">
        <ThreadStateIcon thread={thread} className="mr-2" />
        <span className="text-muted mr-2">#{thread.number}</span> {thread.title}
    </Link>
)
