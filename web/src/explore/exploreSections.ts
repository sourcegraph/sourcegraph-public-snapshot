import { asyncComponent } from '../util/asyncComponent'
import { ExploreSectionDescriptor } from './ExploreArea'

export const exploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    {
        render: asyncComponent(
            () => import('../extensions/explore/ExtensionViewsExploreSection'),
            'ExtensionViewsExploreSection'
        ),
    },
    {
        render: asyncComponent(
            () => import('../integrations/explore/IntegrationsExploreSection'),
            'IntegrationsExploreSection'
        ),
    },
    {
        render: asyncComponent(
            () => import('../repo/explore/RepositoriesExploreSection'),
            'RepositoriesExploreSection'
        ),
    },
    {
        render: asyncComponent(
            () => import('../search/saved-queries/explore/SavedSearchesExploreSection'),
            'SavedSearchesExploreSection'
        ),
    },
    {
        render: asyncComponent(
            () => import('../usageStatistics/explore/SiteUsageExploreSection'),
            'SiteUsageExploreSection'
        ),
        condition: ({ authenticatedUser }) =>
            (!window.context.sourcegraphDotComMode || window.context.debug) && !!authenticatedUser,
    },
]
