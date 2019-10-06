import { NavItemWithIconDescriptor } from '../../util/contributions'
import { CampaignsIcon } from '../expCampaigns/icons'

export const enterpriseNamespaceAreaHeaderNavItems: readonly Pick<
    NavItemWithIconDescriptor,
    Exclude<keyof NavItemWithIconDescriptor, 'condition'>
>[] = [
    {
        to: '/campaigns',
        label: 'Campaigns',
        icon: CampaignsIcon,
    },
]
