import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../../components/time/Timestamp'
import { PersonLink } from '../../../../user/PersonLink'
import { ThreadStateBadge } from '../../../threadlike/threadState/ThreadStateBadge'
import { ThreadStateButton } from '../../form/ThreadStateButton'
import { ThreadSettings } from '../../settings'
import { ThreadStateItemsProgressBar } from '../actions/ThreadStateItemsProgressBar'
import { ThreadHeaderEditableTitle } from '../header/ThreadHeaderEditableTitle'
import { ThreadDescription } from './ThreadDescription'

interface Props extends ExtensionsControllerProps {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    /** The project containing the thread. */
    project: Pick<GQL.IProject, 'id' | 'name' | 'url'> | null

    areaURL: string

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The overview for a single thread.
 */
export const ThreadOverview: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    project,
    areaURL,
    className = '',
    ...props
}) => (
    <div className={`thread-overview ${className || ''}`}>
        <hr className="my-0" />
        <div className="d-flex align-items-center py-3">
            <ThreadStateBadge thread={thread} className="mr-3" />
            <div>
                <small>
                    Opened <Timestamp date={thread.createdAt} /> by{' '}
                    <strong>
                        <PersonLink user={thread.author} />
                    </strong>
                </small>
                {
                    /*thread.type === GQL.ThreadType.ISSUE &&*/ <ThreadStateItemsProgressBar
                        className="mt-1 mb-3"
                        height="0.3rem"
                    />
                }
            </div>
            <ThreadStateButton
                {...props}
                thread={thread}
                onThreadUpdate={onThreadUpdate}
                className="ml-2"
                buttonClassName="btn-link btn-sm"
            />
        </div>
        <hr className="my-0" />
        <ThreadHeaderEditableTitle
            {...props}
            thread={thread}
            onThreadUpdate={onThreadUpdate}
            className="thread-overview__thread-title py-3"
        />
        <ThreadDescription {...props} thread={thread} onThreadUpdate={onThreadUpdate} />
    </div>
)
