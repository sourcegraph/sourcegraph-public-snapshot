import { NotificationType } from '@sourcegraph/extension-api-classes'
import LightbulbIcon from 'mdi-react/LightbulbIcon'
import React, { useCallback, useState } from 'react'
import * as sourcegraph from 'sourcegraph'
import { ActionType } from '../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { useOnActionClickCallback } from '../../../actions/useOnActionClickCallback'
import { ChangesetCreationStatus, createChangeset } from '../../../changesets/preview/backend'
import { NotificationActions } from '../../../notifications/actions/NotificationActions'
import { NotificationTypeIcon } from '../../../notifications/NotificationTypeIcon'
import {
    ChangesetButtonOrLinkExistingChangeset,
    PENDING_CREATION,
} from '../../../tasks/list/item/ChangesetButtonOrLink'

interface Props extends ExtensionsControllerProps {
    notification: sourcegraph.Notification
    className?: string
    contentClassName?: string
}

const LOADING = 'loading' as const

/**
 * A notification associated with a check.
 */
export const CheckNotification: React.FunctionComponent<Props> = ({
    notification,
    className = '',
    contentClassName = '',
    extensionsController,
}) => {
    const [createdChangesetOrLoading, setCreatedChangesetOrLoading] = useState<ChangesetButtonOrLinkExistingChangeset>(
        null
    )
    const onPlanActionClick = useCallback(
        async (plan: ActionType['plan'], creationStatus: ChangesetCreationStatus) => {
            setCreatedChangesetOrLoading(PENDING_CREATION)
            try {
                setCreatedChangesetOrLoading(
                    await createChangeset(
                        { extensionsController },
                        {
                            title: plan.plan.operations[0].command.title,
                            contents: notification.message,
                            status: creationStatus,
                            plan: plan.plan,
                            changesetActionDescriptions: [
                                {
                                    title: plan.plan.title,
                                    timestamp: Date.now(),
                                    user: 'sqs',
                                },
                            ],
                        }
                    )
                )
            } catch (err) {
                setCreatedChangesetOrLoading(null)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error creating changeset: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController, notification.message]
    )
    const isPlanLoading = createdChangesetOrLoading === LOADING

    const [onCommandActionClick, isCommandLoading] = useOnActionClickCallback(extensionsController)

    const disabled = isPlanLoading || isCommandLoading

    return (
        <div className={`status-notification ${className}`}>
            <div className={`d-flex align-items-start ${contentClassName}`}>
                <NotificationTypeIcon type={notification.type} className="icon-inline h3 mb-0 mr-3 flex-0" />
                <div className="flex-1">
                    <div className="d-flex align-items-start justify-content-between">
                        <h3 className="mb-0 font-weight-normal">{notification.message}</h3>
                        <button className="btn btn-success">
                            <LightbulbIcon className="icon-inline mr-2" /> Fix
                        </button>
                    </div>
                    {notification.actions && false && (
                        <NotificationActions
                            actions={notification.actions}
                            onPlanActionClick={onPlanActionClick}
                            onCommandActionClick={onCommandActionClick}
                            existingChangeset={createdChangesetOrLoading}
                            disabled={disabled}
                            className="mt-4"
                        />
                    )}
                </div>
            </div>
        </div>
    )
}
