import { CampaignsIcon } from '../campaigns/icons'
import { OrgAreaHeaderNavItem } from '../../org/area/OrgHeader'

export const enterpriseNamespaceAreaHeaderNavItems: OrgAreaHeaderNavItem[] = [
    {
        to: '/campaigns',
        label: 'Campaigns',
        icon: CampaignsIcon,
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom,
    },
]
