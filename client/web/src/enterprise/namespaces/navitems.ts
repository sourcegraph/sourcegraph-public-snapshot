import { NamespaceAreaNavItem } from '../../namespaces/NamespaceArea'
import { CampaignsIconNamespaceNav } from '../campaigns/icons'

export const enterpriseNamespaceAreaHeaderNavItems: readonly NamespaceAreaNavItem[] = [
    {
        to: '/batch-changes',
        label: 'Batch changes',
        icon: CampaignsIconNamespaceNav,
        condition: ({ isSourcegraphDotCom }: { isSourcegraphDotCom: boolean }): boolean =>
            !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
]
