import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../../components/time/Timestamp'
import { PersonLink } from '../../../../user/PersonLink'
import { CheckThreadActivationStatusButton } from '../../../checks/threads/form/CheckThreadActivationStatusButton'
import { ThreadStatusBadge } from '../../components/threadStatus/ThreadStatusBadge'
import { ThreadStatusButton } from '../../form/ThreadStatusButton'
import { ThreadSettings } from '../../settings'
import { ThreadStatusItemsProgressBar } from '../actions/ThreadStatusItemsProgressBar'
import { ThreadHeaderEditableTitle } from '../header/ThreadHeaderEditableTitle'
import { ThreadBreadcrumbs } from './ThreadBreadcrumbs'
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
        <ThreadBreadcrumbs thread={thread} project={project} areaURL={areaURL} className="py-3" />
        <hr className="my-0" />
        <div className="d-flex align-items-center py-3">
            <ThreadStatusBadge thread={thread} className="mr-3" />
            <div>
                <small>
                    Opened <Timestamp date={thread.createdAt} /> by{' '}
                    <strong>
                        <PersonLink user={thread.author} />
                    </strong>
                </small>
                {thread.type === GQL.ThreadType.CHECK && (
                    <ThreadStatusItemsProgressBar className="mt-1 mb-3" height="0.3rem" />
                )}
            </div>
            <div className="flex-1" />
            {thread.type === GQL.ThreadType.CHECK && thread.status !== GQL.ThreadStatus.CLOSED && (
                <CheckThreadActivationStatusButton
                    {...props}
                    thread={thread}
                    onThreadUpdate={onThreadUpdate}
                    className="ml-2"
                    buttonClassName={`${
                        thread.status === GQL.ThreadStatus.INACTIVE ? 'btn-success' : 'btn-link'
                    } btn-sm`}
                />
            )}
            <ThreadStatusButton
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
