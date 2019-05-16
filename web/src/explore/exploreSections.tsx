import React from 'react'
const ExtensionViewsExploreSection = React.lazy(async () => ({
    default: (await import('../extensions/explore/ExtensionViewsExploreSection')).ExtensionViewsExploreSection,
}))
const IntegrationsExploreSection = React.lazy(async () => ({
    default: (await import('../integrations/explore/IntegrationsExploreSection')).IntegrationsExploreSection,
}))
const RepositoriesExploreSection = React.lazy(async () => ({
    default: (await import('../repo/explore/RepositoriesExploreSection')).RepositoriesExploreSection,
}))
const SiteUsageExploreSection = React.lazy(async () => ({
    default: (await import('../usageStatistics/explore/SiteUsageExploreSection')).SiteUsageExploreSection,
}))
import { ExploreSectionDescriptor } from './ExploreArea'

export const exploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    {
        render: props => <ExtensionViewsExploreSection {...props} />,
    },
    { render: props => <IntegrationsExploreSection {...props} /> },
    {
        render: props => <RepositoriesExploreSection {...props} />,
    },
    {
        render: props => <SiteUsageExploreSection {...props} />,
        condition: ({ authenticatedUser }) =>
            (!window.context.sourcegraphDotComMode || window.context.debug) && !!authenticatedUser,
    },
]
