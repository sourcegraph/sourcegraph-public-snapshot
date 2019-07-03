import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { NotificationTypeIcon } from '../../../../notifications/NotificationTypeIcon'
import { ActionsWithPreview } from '../../../actions/ActionsWithPreview'

interface Props extends ExtensionsControllerProps {
    notification: sourcegraph.Notification
    className?: string
    history: H.History
    location: H.Location
}

/**
 * A notification associated with a status.
 */
export const StatusNotification: React.FunctionComponent<Props> = ({ notification, className = '', ...props }) => (
    <div className={`status-notification card ${className}`}>
        <section className="card-body">
            <h4 className="card-title mb-0 font-weight-normal d-flex align-items-center">
                <NotificationTypeIcon type={notification.type} className="icon-inline mr-2" />
                {notification.message}
            </h4>
        </section>
        {notification.actions && (
            <ActionsWithPreview {...props} actionsOrError={notification.actions}>
                {({ actions, preview }) => (
                    <>
                        {actions}
                        {preview}
                    </>
                )}
            </ActionsWithPreview>
        )}
    </div>
)
