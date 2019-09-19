import React from 'react'
import { CampaignList } from '../../list/CampaignList'
import { useCampaigns } from '../../list/useCampaigns'
import AddIcon from 'mdi-react/AddIcon'
import { Link } from '../../../../../../shared/src/components/Link'

interface Props {}

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = props => {
    const campaigns = useCampaigns()
    return (
        <>
            <h1>Campaigns</h1>
            <p>Track large-scale code changes</p>
            {campaigns && campaigns.nodes.length > 0 && (
                <div className="text-right mb-1">
                    <Link to="/campaigns/new" className="btn btn-primary">
                        <AddIcon className="icon-inline" /> New Campaign
                    </Link>
                </div>
            )}
            <CampaignList {...props} campaigns={campaigns} />
        </>
    )
}
