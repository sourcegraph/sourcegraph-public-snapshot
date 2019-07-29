import { NotificationType } from '@sourcegraph/extension-api-classes'
import CloseIcon from 'mdi-react/CloseIcon'
import React, { useCallback } from 'react'
import { Link } from 'react-router-dom'
import { map, mapTo } from 'rxjs/operators'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../../backend/graphql'
import { ThreadStatusFields } from '../../../threadsOLD/components/threadStatus/threadStatus'
import { ThreadStatusIcon } from '../../../threadsOLD/components/threadStatus/ThreadStatusIcon'

const removeThreadsFromThread = (input: GQL.IRemoveThreadsFromThreadOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation RemoveThreadsFromThread($thread: ID!, $threads: [ID!]!) {
                removeThreadsFromThread(thread: $thread, threads: $threads) {
                    alwaysNil
                }
            }
        `,
        input
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(void 0)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    thread: Pick<GQL.IThread, 'id'>
    thread: Pick<GQL.IDiscussionThread, 'id' | 'title' | 'url'> & ThreadStatusFields
    onUpdate: () => void
}

/**
 * An item in the list of a thread's threads.
 */
export const ThreadThreadListItem: React.FunctionComponent<Props> = ({
    thread,
    thread,
    onUpdate,
    extensionsController,
}) => {
    const onRemoveClick = useCallback(async () => {
        try {
            await removeThreadsFromThread({ thread: thread.id, threads: [thread.id] })
            onUpdate()
        } catch (err) {
            extensionsController.services.notifications.showMessages.next({
                message: `Error removing thread from thread: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [thread.id, extensionsController.services.notifications.showMessages, onUpdate, thread.id])

    return (
        <div className="d-flex align-items-center justify-content-between">
            <Link to={thread.url} className="text-decoration-none">
                <ThreadStatusIcon thread={thread} className="mr-2" />
                {thread.title}
            </Link>
            <button
                className="btn btn-link btn-sm p-1"
                aria-label="Remove thread from thread"
                onClick={onRemoveClick}
            >
                <CloseIcon className="icon-inline" />
            </button>
        </div>
    )
}
