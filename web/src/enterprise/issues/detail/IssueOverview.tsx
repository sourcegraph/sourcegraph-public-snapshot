import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { Comment } from '../../comments/Comment'
import { IssueAreaContext } from './IssueArea'
import { IssueHeaderEditableTitle } from './header/IssueHeaderEditableTitle'

interface Props extends Pick<IssueAreaContext, 'issue' | 'onIssueUpdate'>, ExtensionsControllerProps {
    className?: string

    history: H.History
}

/**
 * The overview for a single issue.
 */
export const IssueOverview: React.FunctionComponent<Props> = ({
    issue,
    onIssueUpdate,
    className = '',
    ...props
}) => (
    <div className={`issue-overview ${className || ''}`}>
        <IssueHeaderEditableTitle
            {...props}
            issue={issue}
            onIssueUpdate={onIssueUpdate}
            className="mb-3"
        />
        <Comment
            {...props}
            comment={issue}
            onCommentUpdate={onIssueUpdate}
            createdVerb="opened issue"
            emptyBody="No description provided."
            className="mb-3"
        />
    </div>
)
