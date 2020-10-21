import { NamespaceAreaNavItem } from '../../namespaces/NamespaceArea'
import { CampaignsIconNav } from '../campaigns/icons'

export const enterpriseNamespaceAreaHeaderNavItems: readonly NamespaceAreaNavItem[] = [
    {
        to: '/campaigns',
        label: 'Campaigns',
        icon: CampaignsIconNav,
        condition: ({ isSourcegraphDotCom }: { isSourcegraphDotCom: boolean }): boolean =>
            !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
]
