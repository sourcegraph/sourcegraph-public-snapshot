import { NamespaceAreaRoute } from '../../../namespaces/NamespaceArea'
import { NavItemWithIconDescriptor } from '../../../util/contributions'
import { lazyComponent } from '../../../util/lazyComponent'
import { GraphIcon } from '../icons'

const COMMON: Pick<NamespaceAreaRoute, 'condition' | 'hideNamespaceAreaSidebar'> = {
    condition: () => window.context?.graphsEnabled,
    hideNamespaceAreaSidebar: false,
}

export const graphOwnerAreaRoutes: readonly NamespaceAreaRoute[] = [
    {
        path: '/graphs',
        exact: true,
        render: lazyComponent(() => import('./GraphOwnerListGraphsPage'), 'GraphOwnerListGraphsPage'),
        ...COMMON,
    },
    {
        path: '/graphs/new',
        exact: true,
        render: lazyComponent(() => import('./GraphOwnerNewGraphPage'), 'GraphOwnerNewGraphPage'),
        ...COMMON,
    },
    {
        path: '/graphs/:graphName',
        render: lazyComponent(() => import('../graphArea/GraphArea'), 'GraphArea'),
        ...COMMON,
    },
]

export const graphOwnerNavItems: readonly NavItemWithIconDescriptor[] = [
    {
        to: '/graphs',
        label: 'Graphs',
        icon: GraphIcon,
        condition: () => window.context?.graphsEnabled,
    },
]
