import { BatchChangesIconNamespaceNav } from '../../batches/icons'
import { CatalogIcon } from '../../catalog'
import { NamespaceAreaNavItem } from '../../namespaces/NamespaceArea'

export const enterpriseNamespaceAreaHeaderNavItems: readonly NamespaceAreaNavItem[] = [
    {
        to: '/batch-changes',
        label: 'Batch Changes',
        icon: BatchChangesIconNamespaceNav,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
    },
    {
        to: '/catalog',
        label: 'Catalog',
        icon: CatalogIcon,
        condition: ({ catalogEnabled }) => catalogEnabled,
    },
]
