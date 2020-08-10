import { userAreaHeaderNavItems } from '../../user/area/navitems'
import { UserAreaHeaderNavItem } from '../../user/area/UserAreaHeader'
import { CampaignsIcon } from '../campaigns/icons'

export const enterpriseUserAreaHeaderNavItems: readonly UserAreaHeaderNavItem[] = [
    ...userAreaHeaderNavItems,
    {
        to: '/campaigns',
        label: 'Campaigns',
        icon: CampaignsIcon,
        condition: ({ isSourcegraphDotCom }): boolean =>
            !isSourcegraphDotCom && window.context.experimentalFeatures.automation === 'enabled',
    },
]
