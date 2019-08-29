import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CommentList } from '../../../comments/CommentList'
import { ThreadTimeline } from '../timeline/ThreadTimeline'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IThread, 'id'>

    className?: string
    history: H.History
}

/**
 * The activity related to an thread.
 */
export const ThreadActivity: React.FunctionComponent<Props> = ({ thread, className = '', ...props }) => (
    <div className={`thread-activity ${className}`}>
        <ThreadTimeline {...props} thread={thread} timelineItemsClassName="pb-6" />
        <CommentList {...props} commentable={thread} />
    </div>
)
