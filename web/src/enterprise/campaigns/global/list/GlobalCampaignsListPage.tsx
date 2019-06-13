import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { CampaignsList } from '../../list/CampaignsList'
import { useCampaigns } from '../../list/useCampaigns'
import { GlobalNewCampaignDropdownButton } from './GlobalNewCampaignDropdownButton'

interface Props extends ExtensionsControllerNotificationProps {}

/**
 * A list of all campaigns.
 */
export const GlobalCampaignsListPage: React.FunctionComponent<Props> = props => {
    const campaigns = useCampaigns()
    return (
        <>
            <GlobalNewCampaignDropdownButton className="mb-3" />
            <CampaignsList {...props} campaigns={campaigns} />
        </>
    )
}
