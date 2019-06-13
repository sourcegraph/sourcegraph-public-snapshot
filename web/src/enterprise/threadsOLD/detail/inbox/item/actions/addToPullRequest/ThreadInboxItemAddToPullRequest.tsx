import SourcePullIcon from 'mdi-react/SourcePullIcon'
import React, { useCallback, useState } from 'react'
import { ButtonDropdown, DropdownToggle } from 'reactstrap'
import { NotificationType } from '../../../../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../../../shared/src/graphql/schema'
import { updateThreadSettings } from '../../../../../../../discussions/backend'
import { PullRequest, ThreadSettings } from '../../../../../settings'
import { PullRequestDropdownMenu } from './PullRequestDropdownMenu'

interface Props extends ExtensionsControllerNotificationProps {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings
    inboxItem: GQL.IDiscussionThreadTargetRepo
    buttonClassName?: string
}

/**
 * An action on a thread inbox item to add the item to an existing or new pull request.
 */
export const ThreadInboxItemAddToPullRequest: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    inboxItem,
    buttonClassName = 'btn-secondary',
    extensionsController,
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const toggleIsOpen = useCallback(() => setIsOpen(!isOpen), [isOpen])

    const onCreateClick = useCallback(async () => {
        try {
            onThreadUpdate(
                await updateThreadSettings(thread, {
                    ...threadSettings,
                    pullRequests: [
                        ...(threadSettings.pullRequests || []),
                        {
                            repo: inboxItem.repository.name,
                            status: 'pending' as const,
                            number: undefined,
                            items: [inboxItem.id],
                        },
                    ],
                })
            )
        } catch (err) {
            extensionsController.services.notifications.showMessages.next({
                message: `Error creating pull request: ${err.message}`,
                type: NotificationType.Error,
            })
        }
    }, [
        onThreadUpdate,
        thread,
        threadSettings,
        inboxItem.repository.name,
        inboxItem.id,
        extensionsController.services.notifications.showMessages,
    ])

    const onAddToExistingClick = useCallback(
        async (_pull: PullRequest) => {
            try {
                onThreadUpdate(
                    await updateThreadSettings(thread, {
                        ...threadSettings,
                        pullRequests: (threadSettings.pullRequests || []).map(pull => {
                            if (pull.repo === inboxItem.repository.name) {
                                return { ...pull, items: [...pull.items, inboxItem.id] }
                            }
                            return pull
                        }),
                    })
                )
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error adding to existing pull request: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [
            onThreadUpdate,
            thread,
            threadSettings,
            inboxItem.repository.name,
            inboxItem.id,
            extensionsController.services.notifications.showMessages,
        ]
    )

    return (
        <ButtonDropdown isOpen={isOpen} toggle={toggleIsOpen}>
            <DropdownToggle className={`btn ${buttonClassName}`} color="none">
                <SourcePullIcon className="icon-inline" /> Add to pull request
            </DropdownToggle>
            <PullRequestDropdownMenu
                threadSettings={threadSettings}
                inboxItem={inboxItem}
                onCreateClick={onCreateClick}
                onAddToExistingClick={onAddToExistingClick}
                right={true}
            />
        </ButtonDropdown>
    )
}
