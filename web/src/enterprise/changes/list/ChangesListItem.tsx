import H from 'history'
import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { SearchResult } from '../../../components/SearchResult'
import { ChangeStatusIcon } from '../components/changeStatus/ChangeStatusIcon'

interface Props {
    change: GQL.ICommitSearchResult
    location: H.Location
    isLightTheme: boolean
}

/**
 * A list item for a change in {@link ChangesList}.
 */
export const ChangesListItem: React.FunctionComponent<Props> = ({ change, location, isLightTheme }) => (
    <li className="list-group-item p-2">
        <div className="d-flex align-items-start justify-content-stretch">
            {/* TODO!(sqs) */}
            {/* <ChangeStatusIcon change={{ status: 123 }} className="small mr-2 mt-2" /> */}
            <SearchResult result={change} isLightTheme={isLightTheme} />
        </div>
    </li>
)
