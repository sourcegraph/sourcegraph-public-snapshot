import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignListItem } from './CampaignListItem'
import { Link } from '../../../../../shared/src/components/Link'

interface Props {
    campaigns?: GQL.ICampaignConnection
}

/**
 * Renders a list of the given campaigns.
 */
export const CampaignList: React.FunctionComponent<Props> = ({ campaigns, ...props }) => (
    <div className="campaign-list">
        {campaigns === undefined ? (
            <LoadingSpinner className="icon-inline mt-3" />
        ) : (
            <>
                {campaigns.nodes.length > 0 ? (
                    <ul className="list-unstyled">
                        {campaigns.nodes.map(campaign => (
                            <li key={campaign.id} className="card p-2 mt-2">
                                <CampaignListItem {...props} campaign={campaign} />
                            </li>
                        ))}
                    </ul>
                ) : (
                    <div className="p-2 text-muted text-center">
                        <p>There are no campaigns yet.</p>
                        <Link to="/campaigns/new" className="btn btn-primary mt-2">
                            Create a campaign
                        </Link>
                    </div>
                )}
            </>
        )}
    </div>
)
