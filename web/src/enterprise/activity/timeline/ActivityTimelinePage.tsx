import H from 'history'
import React from 'react'
import { ActivityAreaContext } from '../global/ActivityArea'
import { ActivityIcon } from '../icons'
import { ActivityTimeline } from './ActivityTimeline'

interface Props extends ActivityAreaContext {
    history: H.History
    location: H.Location
}

/**
 * The activity timeline page.
 */
export const ActivityTimelinePage: React.FunctionComponent<Props> = props => {
    const query = new URLSearchParams(location.search).get('q') || ''
    const onQueryChange = (query: string) => {
        const params = new URLSearchParams(location.search)
        params.set('q', query)
        props.history.push({ search: `${params}` })
    }

    return (
        <div className="activity-timeline-page mt-3 container">
            <h1 className="h4">
                <ActivityIcon className="icon-inline" /> Activity
            </h1>
            <ActivityTimeline {...props} query={query} onQueryChange={onQueryChange} />
        </div>
    )
}
