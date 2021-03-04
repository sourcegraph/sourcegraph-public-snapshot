import { NamespaceAreaNavItem } from '../../namespaces/NamespaceArea'
import { BatchChangesIconNamespaceNav } from '../campaigns/icons'

export const enterpriseNamespaceAreaHeaderNavItems: readonly NamespaceAreaNavItem[] = [
    {
        to: '/batch-changes',
        label: 'Batch changes',
        icon: BatchChangesIconNamespaceNav,
        condition: ({ isSourcegraphDotCom }: { isSourcegraphDotCom: boolean }): boolean =>
            !isSourcegraphDotCom && window.context.batchChangesEnabled,
    },
]
