import React from 'react'
import { PageTitle } from '../../../../components/PageTitle'
import { MinimalCampaign } from '../CampaignArea'

interface Props {
    campaign: Pick<MinimalCampaign, 'id' | 'name' | 'viewerCanAdminister'>
    campaignSpecID: string | null
}

/**
 * A page for updating a campaign.
 */
export const UpdateCampaignPage: React.FunctionComponent<Props> = ({ campaign, campaignSpecID }) => (
    <>
        <PageTitle title={`Update campaign - ${campaign.name}`} />
        <h1>Update campaign</h1>
        Campaign spec ID: {campaignSpecID}
        TODO(sqs)
    </>
)
