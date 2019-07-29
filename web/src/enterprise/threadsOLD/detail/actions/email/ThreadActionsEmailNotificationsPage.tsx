import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { ThreadAreaContext } from '../../ThreadArea'
import { ThreadActionsEmailNotificationRuleForm } from './ThreadActionsEmailNotificationRuleForm'

interface Props extends ThreadAreaContext, ExtensionsControllerProps {
    history: H.History
    location: H.Location
}

/**
 * The page showing email notification actions for a thread.
 */
export const ThreadActionsEmailNotificationsPage: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    ...props
}) => (
    <div>
        <div className="card mb-3">
            <h3 className="card-header">Email notification rules</h3>
            <div className="card-body">
                <ThreadActionsEmailNotificationRuleForm
                    {...props}
                    thread={thread}
                    onThreadUpdate={onThreadUpdate}
                    threadSettings={threadSettings}
                />
            </div>
        </div>
        <div className="card">
            <h3 className="card-header">Log</h3>
            <div className="card-body">No email messages sent yet.</div>
        </div>
    </div>
)
