import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { NamespaceAreaRoute } from '../../namespaces/NamespaceArea'
import { namespaceAreaRoutes } from '../../namespaces/routes'

export const enterpriseNamespaceAreaRoutes: readonly NamespaceAreaRoute[] = [
    ...namespaceAreaRoutes,
    {
        path: '/catalog',
        render: lazyComponent(() => import('../catalog/pages/overview/NamespaceOverviewPage'), 'NamespaceOverviewPage'),
    },
]
