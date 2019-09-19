import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'

interface Props {
    campaign: Pick<GQL.ICampaign, 'name' | 'id' | 'description'>
}

/**
 * An item in the list of campaigns.
 */
export const CampaignListItem: React.FunctionComponent<Props> = ({ campaign }) => (
    <div className="d-flex">
        <CampaignsIcon className="icon-inline mr-2" />
        <div>
            <h3 className="mb-0">
                <Link to={`/campaigns/${campaign.id}`} className="d-flex align-items-center text-decoration-none">
                    {campaign.name}
                </Link>
            </h3>
            <div>{campaign.description}</div>
        </div>
    </div>
)
