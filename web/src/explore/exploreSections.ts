import { lazyComponent } from '../util/lazyComponent'
import { ExploreSectionDescriptor } from './ExploreArea'

export const exploreSections: readonly ExploreSectionDescriptor[] = [
    {
        render: lazyComponent(
            () => import('../integrations/explore/IntegrationsExploreSection'),
            'IntegrationsExploreSection'
        ),
    },
    {
        render: lazyComponent(() => import('../repo/explore/RepositoriesExploreSection'), 'RepositoriesExploreSection'),
    },
    {
        render: lazyComponent(
            () => import('../usageStatistics/explore/SiteUsageExploreSection'),
            'SiteUsageExploreSection'
        ),
        condition: ({ authenticatedUser }) =>
            (!window.context.sourcegraphDotComMode || window.context.debug) && !!authenticatedUser,
    },
]
