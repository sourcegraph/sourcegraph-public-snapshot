import { NamespaceAreaRoute } from '../../namespaces/NamespaceArea'
import { namespaceAreaRoutes } from '../../namespaces/routes'
import { lazyComponent } from '../../util/lazyComponent'

export const enterpriseNamespaceAreaRoutes: readonly NamespaceAreaRoute[] = [
    ...namespaceAreaRoutes,
    {
        path: '/campaigns',
        render: lazyComponent(() => import('../campaigns/namespace/NamespaceCampaignsArea'), 'NamespaceCampaignsArea'),
    },
]
