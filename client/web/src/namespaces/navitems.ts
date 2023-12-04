import { BatchChangesIconNamespaceNav } from '../batches/icons'

import type { NamespaceAreaNavItem } from './NamespaceArea'

export const namespaceAreaHeaderNavItems: readonly NamespaceAreaNavItem[] = [
    {
        to: '/batch-changes',
        label: 'Batch Changes',
        icon: BatchChangesIconNamespaceNav,
        condition: ({ batchChangesEnabled, isCodyApp }) => batchChangesEnabled && !isCodyApp,
    },
]
