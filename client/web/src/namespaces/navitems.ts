import { BatchChangesIconNamespaceNav } from '../batches/icons'
import { isCodyOnlyLicense } from '../util/license'

import type { NamespaceAreaNavItem } from './NamespaceArea'

const disableCodeSearchFeatures = isCodyOnlyLicense()

export const namespaceAreaHeaderNavItems: readonly NamespaceAreaNavItem[] = [
    {
        to: '/batch-changes',
        label: 'Batch Changes',
        icon: BatchChangesIconNamespaceNav,
        condition: ({ batchChangesEnabled }) => !disableCodeSearchFeatures && batchChangesEnabled,
    },
]
