import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import CheckIcon from 'mdi-react/CheckIcon'
import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { pluralize } from '../../../../../shared/src/util/strings'
import { updateThreadSettings } from '../../../discussions/backend'
import { toCreatedPR } from '../detail/actions/pullRequests/CreatePRButton'
import { ThreadSettings } from '../settings'

interface Props {
    thread: Pick<GQL.IDiscussionThread, 'id'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

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

export const ThreadCreatePullRequestsButton: React.FunctionComponent<Props> = ({
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
            setIsLoading(true)
            try {
                onThreadUpdate(
                    await updateThreadSettings(thread, {
                        ...threadSettings,
                        pullRequests: (threadSettings.pullRequests || []).map(toCreatedPR),
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
        [extensionsController.services.notifications.showMessages, onThreadUpdate, thread, threadSettings]
    )
    const count = (threadSettings.pullRequests || []).filter(pull => pull.status === 'pending').length
    return count > 0 ? (
        <button type="button" disabled={isLoading} className={`btn ${buttonClassName} ${className}`} onClick={onClick}>
            {isLoading ? <LoadingSpinner className="icon-inline" /> : <CheckIcon className="icon-inline" />} Create{' '}
            {count} pending {pluralize('PR', count)}
        </button>
    ) : null
}
