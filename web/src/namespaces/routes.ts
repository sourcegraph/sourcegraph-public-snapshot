import { lazyComponent } from '../util/lazyComponent'
import { NamespaceAreaRoute } from './NamespaceArea'

export const namespaceAreaRoutes: ReadonlyArray<NamespaceAreaRoute> = [
    {
        path: '/projects',
        render: lazyComponent(() => import('./projects/NamespaceProjectsPage'), 'NamespaceProjectsPage'),
    },
]
