import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { IssueTimeline } from '../timeline/IssueTimeline'

interface Props extends ExtensionsControllerProps {
    issue: Pick<GQL.IThread, 'id'>

    className?: string
    history: H.History
}

/**
 * The activity related to an issue.
 */
export const IssueActivity: React.FunctionComponent<Props> = ({ issue, className = '', ...props }) => (
    <div className={`issue-activity ${className}`}>
        <IssueTimeline {...props} issue={issue} className="mb-6" />
    </div>
)
