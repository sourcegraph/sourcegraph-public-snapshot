import H from 'history'
import React from 'react'
import { ThreadsAreaContext } from '../ThreadsArea'
import { ThreadsManageThreadsList } from './ThreadsManageThreadsList'

interface Props extends ThreadsAreaContext {
    history: H.History
    location: H.Location
}

/**
 * The Threads management overview page.
 */
export const ThreadsManageOverviewPage: React.FunctionComponent<Props> = props => (
    <div className="Threads-manage-overview-page">
        <ThreadsManageThreadsList {...props} />
    </div>
)
