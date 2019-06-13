import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import BackupRestoreIcon from 'mdi-react/BackupRestoreIcon'
import WindowCloseIcon from 'mdi-react/WindowCloseIcon'
import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { fetchDiscussionThreadAndComments, updateTargetInThread } from '../../../../../discussions/backend'

interface Props {
    inboxItem: Pick<GQL.IDiscussionThreadTargetRepo, 'id' | 'isIgnored'>
    onInboxItemUpdate: (item: GQL.DiscussionThreadTarget) => void
    thread: Pick<GQL.IDiscussionThread, 'id' | 'idWithoutKind'>
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
 * A button that changes the ignored status of an inbox item.
 */
export const ThreadInboxItemIgnoreButton: React.FunctionComponent<Props> = ({
    inboxItem,
    onInboxItemUpdate,
    thread,
    onThreadUpdate,
    className = '',
    buttonClassName = 'btn-secondary',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const updatedItem = await updateTargetInThread({
                    targetID: inboxItem.id,
                    isIgnored: !inboxItem.isIgnored,
                }).toPromise()
                onInboxItemUpdate(updatedItem)
                onThreadUpdate(await fetchDiscussionThreadAndComments(thread.idWithoutKind).toPromise())
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error ${inboxItem.isIgnored ? 'un' : ''}ignoring item: ${err.message}`,
                    type: NotificationType.Error,
                })
            } finally {
                setIsLoading(false)
            }
        },
        [extensionsController.services.notifications.showMessages, inboxItem.id, inboxItem.isIgnored, onInboxItemUpdate, onThreadUpdate, thread.idWithoutKind]
    )
    const Icon = inboxItem.isIgnored ? BackupRestoreIcon : WindowCloseIcon
    return (
        <button type="submit" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <Icon className="icon-inline" />}{' '}
            {inboxItem.isIgnored ? 'Unignore' : 'Ignore'}
        </button>
    )
}
