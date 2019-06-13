import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { SearchResult } from '../../../components/SearchResult'
import { EventStatusIcon } from '../components/eventStatus/EventStatusIcon'

interface Props {
    event: GQL.ICommitSearchResult
    location: H.Location
    isLightTheme: boolean
}

/**
 * A list item for a event in {@link EventsList}.
 */
export const EventsListItem: React.FunctionComponent<Props> = ({ event, location, isLightTheme }) => (
    <li className="list-group-item p-2">
        <div className="d-flex align-items-start justify-content-stretch">
            {/* TODO!(sqs) */}
            {/* <EventStatusIcon event={{ status: 123 }} className="small mr-2 mt-1" /> */}
            <SearchResult result={event} isLightTheme={isLightTheme} />
        </div>
    </li>
)
