import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { ThreadAreaContext } from '../../ThreadArea'
import { ThreadActionsWebhookRuleForm } from './ThreadActionsWebhookRuleForm'

interface Props extends ThreadAreaContext, ExtensionsControllerProps {
    history: H.History
    location: H.Location
}

/**
 * The page showing webhook actions for a thread.
 */
export const ThreadActionsWebhooksPage: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    ...props
}) => (
    <div>
        <div className="card mb-3">
            <h3 className="card-header">Webhooks</h3>
            <div className="card-body">
                <ThreadActionsWebhookRuleForm
                    {...props}
                    thread={thread}
                    onThreadUpdate={onThreadUpdate}
                    threadSettings={threadSettings}
                />
            </div>
        </div>
        <div className="card">
            <h3 className="card-header">Log</h3>
            <div className="card-body">No webhook requests sent yet.</div>
        </div>
    </div>
)
