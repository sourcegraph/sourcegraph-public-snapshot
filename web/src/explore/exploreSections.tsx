import React from 'react'
import { asyncComponent } from '../util/asyncComponent'
import { ExploreSectionDescriptor } from './ExploreArea'

const ExtensionViewsExploreSection = asyncComponent(
    () => import('../extensions/explore/ExtensionViewsExploreSection'),
    'ExtensionViewsExploreSection'
)
const IntegrationsExploreSection = asyncComponent(
    () => import('../integrations/explore/IntegrationsExploreSection'),
    'IntegrationsExploreSection'
)
const RepositoriesExploreSection = asyncComponent(
    () => import('../repo/explore/RepositoriesExploreSection'),
    'RepositoriesExploreSection'
)
const SavedSearchesExploreSection = asyncComponent(
    () => import('../search/saved-queries/explore/SavedSearchesExploreSection'),
    'SavedSearchesExploreSection'
)
const SiteUsageExploreSection = asyncComponent(
    () => import('../usageStatistics/explore/SiteUsageExploreSection'),
    'SiteUsageExploreSection'
)

export const exploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    {
        render: props => <ExtensionViewsExploreSection {...props} />,
    },
    { render: props => <IntegrationsExploreSection {...props} /> },
    {
        render: props => <RepositoriesExploreSection {...props} />,
    },
    {
        render: props => <SavedSearchesExploreSection {...props} />,
    },
    {
        render: props => <SiteUsageExploreSection {...props} />,
        condition: ({ authenticatedUser }) =>
            (!window.context.sourcegraphDotComMode || window.context.debug) && !!authenticatedUser,
    },
]
