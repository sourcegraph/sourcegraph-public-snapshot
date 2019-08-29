import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { TimelineItems } from '../../../timeline/TimelineItems'
import { useThreadTimelineItems } from './useThreadTimelineItems'

interface Props extends ExtensionsControllerProps {
    thread: Pick<GQL.IThread, 'id'>

    className?: string
    timelineItemsClassName?: string
}

const LOADING = 'loading' as const

/**
 * A timeline of events related to the thread.
 */
export const ThreadTimeline: React.FunctionComponent<Props> = ({ thread, className = '', timelineItemsClassName }) => {
    const [timelineItems] = useThreadTimelineItems(thread)
    return (
        <div className={`thread-timeline ${className}`}>
            {timelineItems === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(timelineItems) ? (
                <div className="alert alert-danger">{timelineItems.message}</div>
            ) : timelineItems.totalCount === 0 ? (
                <span className="text-muted">No events.</span>
            ) : (
                <TimelineItems events={timelineItems} className={timelineItemsClassName} />
            )}
        </div>
    )
}
