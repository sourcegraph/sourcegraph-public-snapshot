import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ErrorLike } from '../../../../../../shared/src/util/errors'
import { DiscussionsThread } from '../../../../repo/blob/discussions/DiscussionsThread'
import { ThreadSettings } from '../../settings'

interface Props extends ExtensionsControllerProps {
    thread: GQL.IDiscussionThread
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    threadSettings: ThreadSettings

    className?: string
    history: H.History
    location: H.Location
}

/**
 * The discussion page fragment for a single thread.
 */
export const ThreadDiscussionPage: React.FunctionComponent<Props> = ({
    thread,
    onThreadUpdate,
    threadSettings,
    className = '',
    ...props
}) => (
    <div className={`thread-discussion-page ${className}`}>
        <DiscussionsThread
            {...props}
            threadIDWithoutKind={thread.idWithoutKind}
            commentClassName="border border-top-0"
            skipFirstComment={true}
        />
    </div>
)
