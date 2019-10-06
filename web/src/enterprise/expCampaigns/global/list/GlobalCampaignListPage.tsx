import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { CampaignList } from '../../list/CampaignList'
import { useCampaigns } from '../../list/useCampaigns'
import { GlobalNewCampaignDropdownButton } from './GlobalNewCampaignDropdownButton'

interface Props extends ExtensionsControllerNotificationProps {}

/**
 * A list of all campaigns.
 */
export const GlobalCampaignListPage: React.FunctionComponent<Props> = props => {
    const campaigns = useCampaigns()
    return (
        <>
            <GlobalNewCampaignDropdownButton className="mb-3" />
            <CampaignList {...props} campaigns={campaigns} />
        </>
    )
}
