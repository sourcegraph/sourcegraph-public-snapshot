import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import FlashIcon from 'mdi-react/FlashIcon'
import PauseCircleIcon from 'mdi-react/PauseCircleIcon'
import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { updateThread } from '../../../../discussions/backend'
import { threadNoun } from '../../../threads/util'

interface Props {
    includeNounInLabel?: boolean
    thread: Pick<GQL.IDiscussionThread, 'id' | 'status' | 'type'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    className?: string
    buttonClassName?: string
    extensionsController: {
        services: {
            notifications: {
                showMessages: Pick<
                    ExtensionsControllerProps<
                        'services'
                    >['extensionsController']['services']['notifications']['showMessages'],
                    'next'
                >
            }
        }
    }
}

/**
 * A button that activates or deactivates a check.
 *
 * TODO!(sqs): currently it only sets it archived ("closed")
 * TODO!(sqs): add tests like for ThreadHeaderEditableTitle
 */
export const CheckThreadActivationStatusButton: React.FunctionComponent<Props> = ({
    includeNounInLabel,
    thread,
    onThreadUpdate,
    className = '',
    buttonClassName = 'btn-secondary',
    extensionsController,
}) => {
    const isActive = thread.status === GQL.ThreadStatus.OPEN_ACTIVE
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const updatedThread = await updateThread({ threadID: thread.id, active: !isActive })
                onThreadUpdate(updatedThread)
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error ${
                        thread.status === GQL.ThreadStatus.INACTIVE ? 'activating' : 'deactivating'
                    } check: ${err.message}`,
                    type: NotificationType.Error,
                })
            } finally {
                setIsLoading(false)
            }
        },
        [extensionsController.services.notifications.showMessages, isActive, onThreadUpdate, thread.id, thread.status]
    )
    const Icon = isActive ? PauseCircleIcon : FlashIcon
    return thread.status === GQL.ThreadStatus.CLOSED ? null : (
        <button type="submit" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline mr-2" /> : <Icon className="icon-inline mr-2" />}{' '}
            {isActive ? 'Deactivate' : 'Activate'} {includeNounInLabel && threadNoun(thread.type)}
        </button>
    )
}
