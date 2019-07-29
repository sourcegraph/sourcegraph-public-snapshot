import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { ThreadAreaContext } from '../../ThreadArea'
import { ThreadActionsSlackNotificationRuleForm } from './ThreadActionsSlackNotificationRuleForm'

interface Props extends ThreadAreaContext, ExtensionsControllerProps {
    history: H.History
    location: H.Location
}

/**
 * The page showing Slack notification actions for a thread.
 */
export const ThreadActionsSlackNotificationsPage: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    ...props
}) => (
    <div>
        <div className="card mb-3">
            <h3 className="card-header">Slack notification rules</h3>
            <div className="card-body">
                <ThreadActionsSlackNotificationRuleForm
                    {...props}
                    thread={thread}
                    onThreadUpdate={onThreadUpdate}
                    threadSettings={threadSettings}
                />
            </div>
        </div>
        <div className="card">
            <h3 className="card-header">Log</h3>
            <div className="card-body">No Slack messages sent yet.</div>
        </div>
    </div>
)
