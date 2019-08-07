import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
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
        <ThreadTimeline {...props} thread={thread} className="mb-6" />
    </div>
)
