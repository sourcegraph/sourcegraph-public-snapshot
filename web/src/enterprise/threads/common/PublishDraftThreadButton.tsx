import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useState } from 'react'
import { map, mapTo } from 'rxjs/operators'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'

const publishDraftThread = (args: GQL.IPublishDraftThreadOnMutationArguments): Promise<void> =>
    mutateGraphQL(
        gql`
            mutation PublishDraftThread($thread: ID!) {
                publishDraftThread(thread: $thread) {
                    id
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
    thread: Pick<GQL.IThread, 'id' | 'kind'>
    onComplete?: () => void
    className?: string
    buttonClassName?: string
}

/**
 * A button that publishes a draft thread.
 */
export const PublishDraftThreadButton: React.FunctionComponent<Props> = ({
    thread,
    onComplete,
    className = '',
    buttonClassName = 'btn-link text-decoration-none',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                await publishDraftThread({ thread: thread.id })
                setIsLoading(false)
                if (onComplete) {
                    onComplete()
                }
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error publishing draft thread: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, onComplete, thread.id]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading && <LoadingSpinner className="icon-inline" />} Publish {thread.kind.toLowerCase()}
        </button>
    )
}
