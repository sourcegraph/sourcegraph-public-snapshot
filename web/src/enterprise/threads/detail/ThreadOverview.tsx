import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { Comment } from '../../comments/Comment'
import { ThreadAreaContext } from './ThreadArea'
import { ThreadHeaderEditableTitle } from './header/ThreadHeaderEditableTitle'

interface Props extends Pick<ThreadAreaContext, 'thread' | 'onThreadUpdate'>, ExtensionsControllerProps {
    className?: string

    history: H.History
}

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
        <ThreadHeaderEditableTitle
            {...props}
            thread={thread}
            onThreadUpdate={onThreadUpdate}
            className="mb-3"
        />
        <Comment
            {...props}
            comment={thread}
            onCommentUpdate={onThreadUpdate}
            createdVerb="opened thread"
            emptyBody="No description provided."
            className="mb-3"
        />
    </div>
)
