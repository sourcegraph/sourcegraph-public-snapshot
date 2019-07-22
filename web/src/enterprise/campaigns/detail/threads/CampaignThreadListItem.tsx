import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { ThreadStatusFields } from '../../../threads/components/threadStatus/threadStatus'
import { ThreadStatusIcon } from '../../../threads/components/threadStatus/ThreadStatusIcon'

interface Props {
    thread: Pick<GQL.IDiscussionThread, 'title' | 'url'> & ThreadStatusFields
}

/**
 * An item in the list of a campaign's threads.
 */
export const CampaignThreadListItem: React.FunctionComponent<Props> = ({ thread }) => (
    <div className="d-flex align-items-center justify-content-between">
        <Link to={thread.url} className="text-decoration-none">
            <ThreadStatusIcon thread={thread} className="mr-2" />
            {thread.title}
        </Link>
    </div>
)
