import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../../shared/src/extensions/controller'
import { ThreadAreaContext } from '../../ThreadArea'
import { ThreadActionsCommitStatusRuleForm } from './ThreadActionsCommitStatusRuleForm'

interface Props extends ThreadAreaContext, ExtensionsControllerProps {
    history: H.History
    location: H.Location
}

/**
 * The page showing commit status actions for a thread.
 */
export const ThreadActionsCommitStatusesPage: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    ...props
}) => (
    <div>
        <div className="card mb-3">
            <h3 className="card-header">Commit status rules</h3>
            <div className="card-body">
                <ThreadActionsCommitStatusRuleForm
                    {...props}
                    thread={thread}
                    onThreadUpdate={onThreadUpdate}
                    threadSettings={threadSettings}
                />
            </div>
        </div>
        <div className="card">
            <h3 className="card-header">Log</h3>
            <div className="card-body">No commit statuses found.</div>
        </div>
    </div>
)
