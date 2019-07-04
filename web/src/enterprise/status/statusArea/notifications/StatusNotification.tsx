import { NotificationType } from '@sourcegraph/extension-api-classes'
import React, { useCallback, useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { ActionType } from '../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { useOnActionClickCallback } from '../../../actions/useOnActionClickCallback'
import { ChangesetCreationStatus, createChangeset } from '../../../changesets/preview/backend'
import { NotificationActions } from '../../../notifications/actions/NotificationActions'
import { NotificationTypeIcon } from '../../../notifications/NotificationTypeIcon'

interface Props extends ExtensionsControllerProps {
    notification: sourcegraph.Notification
    className?: string
    contentClassName?: string
}

const LOADING = 'loading' as const

/**
 * A notification associated with a status.
 */
export const StatusNotification: React.FunctionComponent<Props> = ({
    notification,
    className = '',
    contentClassName = '',
    extensionsController,
}) => {
    const [createdChangesetOrLoading, setCreatedChangesetOrLoading] = useState<
        typeof LOADING | Pick<GQL.IDiscussionThread, 'idWithoutKind' | 'url' | 'status'>
    >()
    const onPlanActionClick = useCallback(
        async (plan: ActionType['plan'], creationStatus: ChangesetCreationStatus) => {
            setCreatedChangesetOrLoading(LOADING)
            try {
                setCreatedChangesetOrLoading(
                    await createChangeset({
                        title: plan.plan.operations[0].command.title,
                        contents: notification.message,
                        status: creationStatus,
                        changesetActionDescriptions: [
                            {
                                title: plan.plan.operations[0].command.title,
                                timestamp: Date.now(),
                                user: 'sqs',
                            },
                        ],
                    })
                )
            } catch (err) {
                setCreatedChangesetOrLoading(undefined)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error creating changeset: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, notification.message]
    )
    const isPlanLoading = createdChangesetOrLoading === LOADING

    const [onCommandActionClick, isCommandLoading] = useOnActionClickCallback(extensionsController)

    const disabled = isPlanLoading || isCommandLoading

    return (
        <div className={`status-notification ${className}`}>
            <section className={contentClassName}>
                <h4 className="mb-0 font-weight-normal d-flex align-items-center">
                    <NotificationTypeIcon type={notification.type} className="icon-inline mr-2" />
                    {notification.message}
                </h4>
            </section>
            {notification.actions && (
                <div className={contentClassName}>
                    <NotificationActions
                        actions={notification.actions}
                        onPlanActionClick={onPlanActionClick}
                        onCommandActionClick={onCommandActionClick}
                        disabled={disabled}
                        className="mt-4"
                    />
                </div>
            )}
        </div>
    )
}
