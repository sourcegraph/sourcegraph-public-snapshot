import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { TimelineItems } from '../../../timeline/TimelineItems'
import { useIssueTimelineItems } from './useIssueTimelineItems'

interface Props extends ExtensionsControllerProps {
    issue: Pick<GQL.IIssue, 'id'>

    className?: string
}

const LOADING = 'loading' as const

/**
 * A timeline of events related to the issue.
 */
export const IssueTimeline: React.FunctionComponent<Props> = ({ issue, className = '' }) => {
    const [timelineItems] = useIssueTimelineItems(issue)
    return (
        <div className={`issue-timeline ${className}`}>
            {timelineItems === LOADING ? (
                <LoadingSpinner className="icon-inline" />
            ) : isErrorLike(timelineItems) ? (
                <div className="alert alert-danger">{timelineItems.message}</div>
            ) : timelineItems.totalCount === 0 ? (
                <span className="text-muted">No events.</span>
            ) : (
                <TimelineItems events={timelineItems} />
            )}
        </div>
    )
}
