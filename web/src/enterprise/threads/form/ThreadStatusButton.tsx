import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import BackupRestoreIcon from 'mdi-react/BackupRestoreIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { updateThread } from '../../../discussions/backend'
import { threadNoun } from '../util'

interface Props extends ExtensionsControllerNotificationProps {
    includeNounInLabel?: boolean
    thread: Pick<GQL.IDiscussionThread, 'id' | 'status' | 'type'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    className?: string
    buttonClassName?: string
}

/**
 * A button that changes the status of a thread.
 *
 * TODO!(sqs): currently it only sets it archived ("closed")
 * TODO!(sqs): add tests like for ThreadHeaderEditableTitle
 */
export const ThreadStatusButton: React.FunctionComponent<Props> = ({
    includeNounInLabel,
    thread,
    onThreadUpdate,
    className = '',
    buttonClassName = 'btn-secondary',
    extensionsController,
}) => {
    const isOpen = thread.status !== GQL.ThreadStatus.CLOSED
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                // Include `active: false` so that reopening a check doesn't immediately restart all
                // of its actions (which is probably undesirable).
                const updatedThread = await updateThread({
                    threadID: thread.id,
                    status:
                        thread.status === GQL.ThreadStatus.OPEN_ACTIVE
                            ? GQL.ThreadStatus.CLOSED
                            : GQL.ThreadStatus.OPEN_ACTIVE,
                })
                onThreadUpdate(updatedThread)
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error archiving thread: ${err.message}`,
                    type: NotificationType.Error,
                })
            } finally {
                setIsLoading(false)
            }
        },
        [thread.id, thread.status, onThreadUpdate, extensionsController.services.notifications.showMessages]
    )
    const Icon = isOpen ? CheckIcon : BackupRestoreIcon
    return (
        <button type="submit" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <Icon className="icon-inline" />}{' '}
            {isOpen ? 'Close' : 'Reopen'} {includeNounInLabel && threadNoun(thread.type)}
        </button>
    )
}
