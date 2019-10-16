import React from 'react'
import AddIcon from 'mdi-react/AddIcon'
import { Link } from '../../../../../../shared/src/components/Link'
import { useCampaigns } from '../../list/useCampaigns'
import { CampaignList } from '../../list/CampaignList'

interface Props {}

/**
 * A list of all campaigns on the Sourcegraph instance.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = () => {
    const campaigns = useCampaigns()
    return (
        <>
            <h1>Campaigns</h1>
            <p>Track large-scale code changes</p>

            <div className="text-right mb-1">
                <Link to="/campaigns/new" className="btn btn-primary">
                    <AddIcon className="icon-inline" /> New Campaign
                </Link>
            </div>
            <CampaignList campaigns={campaigns} />
        </>
    )
}
