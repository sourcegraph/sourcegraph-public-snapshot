import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { ThreadAreaContext } from '../../ThreadArea'
import { ThreadActionsWebhookRuleForm } from './ThreadActionsWebhookRuleForm'

interface Props extends ThreadAreaContext, ExtensionsControllerProps {
    history: H.History
    location: H.Location
}

/**
 * The page showing editor integration information for a thread.
 */
export const ThreadActionsEditorPage: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    ...props
}) => (
    <div>
        <div className="card mb-3">
            <h3 className="card-header">Editor integration</h3>
            <div className="card-body">You can see this check's results and actions on code in your editor.</div>
            <div className="card-body">
                <a href="https://docs.sourcegraph.com/integration/editor" target="_blank" className="btn btn-primary">
                    Install Sourcegraph editor integrations
                </a>
            </div>
        </div>
    </div>
)
