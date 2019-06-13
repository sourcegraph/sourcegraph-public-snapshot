import AddIcon from 'mdi-react/AddIcon'
import React, { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { fetchDiscussionThreads } from '../../../../discussions/backend'

interface Props {
    className?: string
}

const LOADING: 'loading' = 'loading'

/**
 * A dropdown menu with a list of navigation links related to threads.
 */
export const ThreadsNavItemDropdownMenu: React.FunctionComponent<Props> = ({ className = '' }) => {
    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )

    // tslint:disable-next-line: no-floating-promises
    useMemo(async () => {
        try {
            setThreadsOrError(await fetchDiscussionThreads({}).toPromise())
        } catch (err) {
            setThreadsOrError(asError(err))
        }
    }, [])

    const MAX_THREADS = 5 // TODO!(sqs): hack

    return (
        <DropdownMenu right={true} className={className}>
            <Link to="/threads/-/new" className="dropdown-item d-flex align-items-center">
                <AddIcon className="icon-inline mr-1" /> Start new thread
            </Link>
            <DropdownItem divider={true} />
            {threadsOrError === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading threads...
                </DropdownItem>
            ) : (
                !isErrorLike(threadsOrError) && (
                    <>
                        <DropdownItem header={true} className="py-1">
                            Recent threads
                        </DropdownItem>
                        {threadsOrError.nodes.slice(0, MAX_THREADS).map(thread => (
                            <Link key={thread.id} to={thread.url} className="dropdown-item text-truncate">
                                <small className="text-muted">#{thread.idWithoutKind}</small> {thread.title}
                            </Link>
                        ))}
                    </>
                )
            )}
        </DropdownMenu>
    )
}
