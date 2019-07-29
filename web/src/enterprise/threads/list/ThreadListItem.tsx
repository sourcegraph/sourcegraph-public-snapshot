import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadsIcon } from '../icons'

interface Props {
    thread: Pick<GQL.IThread, 'idInRepository' | 'title' | 'url'>
}

/**
 * An item in the list of threads.
 */
export const ThreadListItem: React.FunctionComponent<Props> = ({ thread }) => (
    <Link to={thread.url} className="d-flex align-items-center text-decoration-none">
        <ThreadsIcon className="icon-inline mr-2" /> <span className="text-muted mr-2">#{thread.idInRepository}</span>{' '}
        {thread.title}
    </Link>
)
