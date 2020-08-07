import { NavItemWithIconDescriptor } from '../../util/contributions'
import { CampaignsIcon } from '../campaigns/icons'

export const enterpriseNamespaceAreaHeaderNavItems: readonly NavItemWithIconDescriptor<{
    isSourcegraphDotCom: boolean
}>[] = [
    {
        to: '/campaigns',
        label: 'Campaigns',
        icon: CampaignsIcon,
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom,
    },
]
