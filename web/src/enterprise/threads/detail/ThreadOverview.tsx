import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { Comment } from '../../comments/Comment'
import { ThreadHeaderEditableTitle } from './header/ThreadHeaderEditableTitle'
import { ThreadAreaContext } from './ThreadArea'

interface Props extends Pick<ThreadAreaContext, 'thread' | 'onThreadUpdate'>, ExtensionsControllerProps {
    className?: string

    history: H.History
}

export const THREAD_COMMENT_CREATED_VERB = 'opened thread'

export const THREAD_COMMENT_EMPTY_BODY = 'No description provided.'

/**
 * The overview for a single thread.
 */
export const ThreadOverview: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    className = '',
    ...props
}) => (
    <div className={`thread-overview ${className || ''}`}>
        <ThreadHeaderEditableTitle {...props} thread={thread} onThreadUpdate={onThreadUpdate} className="mb-3" />
        <Comment
            {...props}
            comment={thread}
            onCommentUpdate={onThreadUpdate}
            createdVerb={THREAD_COMMENT_CREATED_VERB}
            emptyBody={THREAD_COMMENT_EMPTY_BODY}
            className="mb-3"
        />
    </div>
)
