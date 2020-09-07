import { orgAreaHeaderNavItems } from '../../org/area/navitems'
import { OrgAreaHeaderNavItem } from '../../org/area/OrgHeader'
import { enterpriseNamespaceAreaHeaderNavItems } from '../namespaces/navitems'
import { CampaignsIcon } from '../campaigns/icons'

export const enterpriseOrgAreaHeaderNavItems: readonly OrgAreaHeaderNavItem[] = [
    ...orgAreaHeaderNavItems,
    ...enterpriseNamespaceAreaHeaderNavItems,
    {
        to: '/campaigns',
        label: 'Campaigns',
        icon: CampaignsIcon,
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
]
