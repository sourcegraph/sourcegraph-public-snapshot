import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import DeleteIcon from 'mdi-react/DeleteIcon'
import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { updateThread } from '../../../discussions/backend'
import { threadNoun } from '../util'

interface Props extends ExtensionsControllerNotificationProps {
    textLabel?: boolean
    includeNounInLabel?: boolean
    thread: Pick<GQL.IDiscussionThread, 'id' | 'type'>
    history: H.History
    className?: string
    buttonClassName?: string
}

/**
 * A button that permanently deletes a thread.
 */
export const ThreadDeleteButton: React.FunctionComponent<Props> = ({
    textLabel = true,
    includeNounInLabel,
    thread: { id: threadID, type },
    history,
    className = '',
    buttonClassName = 'btn-danger',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            if (!confirm(`Are you sure you want to permanently delete this ${threadNoun(type)}?`)) {
                return
            }
            setIsLoading(true)
            try {
                await updateThread({ threadID, delete: true })
                setIsLoading(false)
                history.push(type === GQL.ThreadType.CHECK ? '/checks' : '/threads')
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error deleting thread: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, history, threadID, type]
    )
    return (
        <button
            type="button"
            disabled={isLoading}
            className={`btn ${buttonClassName} ${className}`}
            onClick={onClick}
            data-tooltip={textLabel ? '' : `Delete ${threadNoun(type)}`}
        >
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <DeleteIcon className="icon-inline" />}{' '}
            {textLabel && <>Delete {includeNounInLabel && threadNoun(type)}</>}
        </button>
    )
}
