import React from 'react'
import { isDiscussionsEnabled } from '../discussions'
import { DiscussionsExploreSection } from '../discussions/explore/DiscussionsExploreSection'
import { ExtensionViewsExploreSection } from '../extensions/explore/ExtensionViewsExploreSection'
import { IntegrationsExploreSection } from '../integrations/explore/IntegrationsExploreSection'
import { RepositoriesExploreSection } from '../repo/explore/RepositoriesExploreSection'
import { SavedSearchesExploreSection } from '../search/saved-queries/explore/SavedSearchesExploreSection'
import { SiteUsageExploreSection } from '../usageStatistics/explore/SiteUsageExploreSection'
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
        render: props => <DiscussionsExploreSection {...props} />,
        condition: ({ configurationCascade }) => isDiscussionsEnabled(configurationCascade),
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
