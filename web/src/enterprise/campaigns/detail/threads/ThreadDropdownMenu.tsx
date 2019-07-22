import React, { useEffect, useState } from 'react'
import { DropdownItem, DropdownMenu } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { fetchDiscussionThreads } from '../../../../discussions/backend'

interface Props {
    /** Called when the user clicks on a thread in the menu. */
    onThreadClick: (thread: Pick<GQL.IDiscussionThread, 'id'>) => void
}

const LOADING: 'loading' = 'loading'

/**
 * A dropdown menu with a list of threads.
 */
export const ThreadDropdownMenu: React.FunctionComponent<Props> = ({ onThreadClick, ...props }) => {
    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )
    useEffect(() => {
        const do2 = async () => {
            try {
                setThreadsOrError(await fetchDiscussionThreads({}).toPromise())
            } catch (err) {
                setThreadsOrError(asError(err))
            }
        }
        // tslint:disable-next-line: no-floating-promises
        do2()
    }, [])

    const MAX_THREADS = 9 // TODO!(sqs): hack

    return (
        <DropdownMenu {...props}>
            <DropdownItem divider={true} />
            {threadsOrError === LOADING ? (
                <DropdownItem header={true} className="py-1">
                    Loading threads...
                </DropdownItem>
            ) : isErrorLike(threadsOrError) ? (
                <DropdownItem header={true} className="py-1">
                    Error loading existing threads
                </DropdownItem>
            ) : (
                <>
                    <DropdownItem header={true} className="py-1">
                        Attach to existing thread...
                    </DropdownItem>
                    {threadsOrError.nodes.slice(0, MAX_THREADS).map(thread => (
                        // tslint:disable-next-line: jsx-no-lambda
                        <DropdownItem key={thread.id} onClick={() => onThreadClick(thread)}>
                            <small className="text-muted">#{thread.idWithoutKind}</small> {thread.title}
                        </DropdownItem>
                    ))}
                </>
            )}
        </DropdownMenu>
    )
}
