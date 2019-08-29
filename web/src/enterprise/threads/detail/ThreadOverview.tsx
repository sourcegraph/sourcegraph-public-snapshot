import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Comment } from '../../comments/Comment'
import { ThreadHeaderEditableTitle } from './header/ThreadHeaderEditableTitle'
import { ThreadAreaContext } from './ThreadArea'
import { Timeline } from '../../../components/timeline/Timeline'
import { GitPullRequestIcon } from '../../../util/octicons'
import { IsDraftTimelineBox } from '../../campaigns/common/IsDraftTimelineBox'
import { PublishDraftThreadButton } from '../common/PublishDraftThreadButton'

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
        <Timeline className="align-items-stretch mb-4">
            <Comment
                {...props}
                comment={thread}
                onCommentUpdate={onThreadUpdate}
                createdVerb={THREAD_COMMENT_CREATED_VERB}
                emptyBody={THREAD_COMMENT_EMPTY_BODY}
            />
            {thread.isDraft && (
                <IsDraftTimelineBox
                    noun={thread.kind.toLowerCase()}
                    action={
                        <PublishDraftThreadButton
                            {...props}
                            thread={thread}
                            onComplete={onThreadUpdate}
                            buttonClassName="btn-secondary"
                        />
                    }
                />
            )}
            {thread.kind === GQL.ThreadKind.CHANGESET && thread.baseRef && thread.headRef && (
                <div className="d-flex align-items-start bg-body border mt-5 p-4 position-relative">
                    <GitPullRequestIcon className="icon-inline mb-0 mr-3" />
                    <div>
                        Request to merge into <code className="bg-secondary py-1 px-2">{thread.baseRef}</code> from{' '}
                        <code className="bg-secondary py-1 px-2">{thread.headRef}</code>
                    </div>
                </div>
            )}
        </Timeline>
    </div>
)
