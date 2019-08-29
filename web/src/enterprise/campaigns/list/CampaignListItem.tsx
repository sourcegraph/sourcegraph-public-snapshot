import React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { CampaignsIcon } from '../icons'

interface Props {
    campaign: Pick<GQL.ICampaign, 'name' | 'url'>
}

/**
 * An item in the list of campaigns.
 */
export const CampaignListItem: React.FunctionComponent<Props> = ({ campaign }) => (
    <div className="d-flex align-items-center justify-content-between">
        <h3 className="mb-0">
            <Link to={campaign.url} className="d-flex align-items-center text-decoration-none">
                <CampaignsIcon className="icon-inline small mr-2" /> {campaign.name}
            </Link>
        </h3>
    </div>
)
