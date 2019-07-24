import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { CampaignsList } from '../../list/CampaignsList'
import { useCampaigns } from '../../list/useCampaignsDefinedInNamespace'

interface Props extends ExtensionsControllerNotificationProps {}

/**
 * A list of all campaigns.
 */
export const GlobalCampaignsListPage: React.FunctionComponent<Props> = props => {
    const campaigns = useCampaigns()
    return <CampaignsList {...props} campaigns={campaigns} />
}
