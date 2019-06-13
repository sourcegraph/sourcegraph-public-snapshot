import H from 'history'
import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { FilteredConnection } from '../../../../components/FilteredConnection'
import { fetchDiscussionThreads } from '../../../../discussions/backend'
import { ThreadsManageThreadsListItem } from './ThreadsManageThreadsListItem'

interface Props {
    history: H.History
    location: H.Location
}

/**
 * The list of threads in the threads management area.
 */
export const ThreadsManageThreadsList: React.FunctionComponent<Props> = ({ history, location }) => (
    <div className="threads-manage-threads-list">
        <FilteredConnection<GQL.IDiscussionThread>
            listClassName="list-group list-group-flush"
            listComponent="ul"
            noun="thread"
            pluralNoun="threads"
            queryConnection={fetchDiscussionThreads}
            nodeComponent={ThreadsManageThreadsListItem}
            hideSearch={false}
            noSummaryIfAllNodesVisible={true}
            history={history}
            location={location}
        />
    </div>
)
