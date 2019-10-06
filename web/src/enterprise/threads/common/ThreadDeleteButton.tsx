import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'

const deleteThread = (args: GQL.IDeleteThreadOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation DeleteThread($thread: ID!) {
                deleteThread(thread: $thread) {
                    alwaysNil
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            mapTo(undefined)
        )
        .toPromise()

interface Props extends ExtensionsControllerNotificationProps {
    thread: Pick<GQL.IThread, 'id'>
    onDelete?: () => void
    className?: string
    buttonClassName?: string
}

/**
 * A button that permanently deletes a thread.
 */
export const ThreadDeleteButton: React.FunctionComponent<Props> = ({
    thread,
    onDelete,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            if (!confirm('Are you sure? Deleting will remove all data associated with the thread.')) {
                return
            }
            setIsLoading(true)
            try {
                await deleteThread({ thread: thread.id })
                setIsLoading(false)
                if (onDelete) {
                    onDelete()
                }
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error deleting thread: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, onDelete, thread.id]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <DeleteIcon className="icon-inline" />} Delete
            thread
        </button>
    )
}
