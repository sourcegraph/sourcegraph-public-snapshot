import { NamespaceAreaNavItem } from '../../namespaces/NamespaceArea'
import { CampaignsIcon } from '../campaigns/icons'
import { graphOwnerNavItems } from '../graphs/graphOwner/routes'

export const enterpriseNamespaceAreaHeaderNavItems: readonly NamespaceAreaNavItem[] = [
    {
        to: '/campaigns',
        label: 'Campaigns',
        icon: CampaignsIcon,
        condition: ({ isSourcegraphDotCom }: { isSourcegraphDotCom: boolean }): boolean =>
            !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
    ...graphOwnerNavItems,
]
