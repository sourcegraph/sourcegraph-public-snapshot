import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Timestamp } from '../../../../components/time/Timestamp'
import { PersonLink } from '../../../../user/PersonLink'
import { CheckThreadActivationStatusButton } from '../../../checks/threads/form/CheckThreadActivationStatusButton'
import { ThreadStatusBadge } from '../../../threads/components/threadStatus/ThreadStatusBadge'
import { ThreadHeaderEditableTitle } from '../../../threads/detail/header/ThreadHeaderEditableTitle'
import { ThreadBreadcrumbs } from '../../../threads/detail/overview/ThreadBreadcrumbs'
import { ThreadDescription } from '../../../threads/detail/overview/ThreadDescription'
import { ThreadStatusButton } from '../../../threads/form/ThreadStatusButton'
import { ThreadSettings } from '../../../threads/settings'

interface Props extends ExtensionsControllerProps {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    areaURL: string

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The overview for a single changeset.
 */
export const ChangesetOverview: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    areaURL,
    className = '',
    ...props
}) => (
    // TODO!(sqs): uses style from other component
    <div className={`thread-overview ${className || ''}`}>
        <ThreadBreadcrumbs thread={thread} project={null /* TODO!(sqs) */} areaURL={areaURL} className="py-3" />
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
            </div>
            <div className="flex-1" />
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
            className="thread-overview__thread-title py-3" // TODO!(sqs): uses style from other component
        />
        <ThreadDescription {...props} thread={thread} onThreadUpdate={onThreadUpdate} />
    </div>
)
