import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThreadsIcon } from '../icons'

interface Props {
    thread: Pick<GQL.IThread, 'name' | 'url'>
}

/**
 * An item in the list of threads.
 */
export const ThreadListItem: React.FunctionComponent<Props> = ({ thread }) => (
    <div className="d-flex align-items-center justify-content-between">
        <h3 className="mb-0">
            <Link to={thread.url} className="text-decoration-none">
                <ThreadsIcon className="icon-inline" /> {thread.name}
            </Link>
        </h3>
    </div>
)
