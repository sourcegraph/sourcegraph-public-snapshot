import H from 'history'
import React, { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { DropdownItem, DropdownMenu, DropdownMenuProps } from 'reactstrap'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { fetchDiscussionThreads } from '../../../../discussions/backend'
import { urlForNewThreadAtSelection } from '../../../../repo/blob/discussions/DiscussionsGutterOverlay'
import { useEffectAsync } from '../../../../util/useEffectAsync'

interface Props extends Pick<DropdownMenuProps, 'right'> {
    /** Called when the user clicks on an existing thread to add to. */
    onAddToExistingThreadClick: (thread: Pick<GQL.IDiscussionThread, 'id'>) => void

    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * A dropdown menu with a list of threads and an option to create a new thread.
 */
export const ThreadDropdownMenu: React.FunctionComponent<Props> = ({
    onAddToExistingThreadClick,
    location,
    ...props
}) => {
    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )
    useEffectAsync(async () => {
        try {
            setThreadsOrError(await fetchDiscussionThreads({}).toPromise())
        } catch (err) {
            setThreadsOrError(asError(err))
        }
    }, [])

    const MAX_THREADS = 9 // TODO!(sqs): hack

    return (
        <DropdownMenu {...props}>
            <Link to={urlForNewThreadAtSelection(location)} className="dropdown-item">
                Start new thread
            </Link>
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
                        <DropdownItem key={thread.id} onClick={() => onAddToExistingThreadClick(thread)}>
                            <small className="text-muted">#{thread.idWithoutKind}</small> {thread.title}
                        </DropdownItem>
                    ))}
                </>
            )}
        </DropdownMenu>
    )
}
