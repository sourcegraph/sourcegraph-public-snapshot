import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../../../shared/src/api/client/services/notifications'
import {
    ExtensionsControllerNotificationProps,
    ExtensionsControllerProps,
} from '../../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../../shared/src/graphql/schema'
import { updateThreadSettings } from '../../../../../discussions/backend'
import { PullRequest, ThreadSettings } from '../../../settings'

export const toCreatedPR = (pull: PullRequest): PullRequest => ({
    ...pull,
    commentsCount: Math.ceil(Math.random() * 17),
    number: Math.ceil(Math.random() * 1000),
    status: 'open' as const,
    title: 'My PR',
    updatedAt: new Date().toISOString(),
    updatedBy: 'sqs',
})

interface Props extends ExtensionsControllerNotificationProps {
    pull: PullRequest
    thread: Pick<GQL.IDiscussionThread, 'id'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    className?: string
    buttonClassName?: string
}

/**
 * A button that creates the PR for a thread.
 */
export const CreatePRButton: React.FunctionComponent<Props> = ({
    pull,
    thread,
    onThreadUpdate,
    threadSettings,
    className = '',
    buttonClassName = 'btn-success',
    extensionsController,
}) => {
    const [isLoading, setIsLoading] = useState(false)
    const onClick = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                onThreadUpdate(
                    await updateThreadSettings(thread, {
                        ...threadSettings,
                        pullRequests: (threadSettings.pullRequests || []).map(p => {
                            if (pull.repo === p.repo) {
                                return toCreatedPR(p)
                            }
                            return p
                        }),
                    })
                )
            } catch (err) {
                extensionsController.services.notifications.showMessages.next({
                    message: `Error creating PR: ${err.message}`,
                    type: NotificationType.Error,
                })
            } finally {
                setIsLoading(false)
            }
        },
        [onThreadUpdate, thread, threadSettings, pull.repo, extensionsController.services.notifications.showMessages]
    )
    return (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <CheckIcon className="icon-inline" />} Create PR
        </button>
    )
}
