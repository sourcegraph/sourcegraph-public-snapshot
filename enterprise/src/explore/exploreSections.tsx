import React from 'react'
import { ExploreSectionDescriptor } from '../../../src/explore/ExploreArea'
import { exploreSections } from '../../../src/explore/exploreSections'
import { ExtensionsExploreSection } from '../extensions/explore/ExtensionsExploreSection'
import { IntegrationsExploreSection } from '../integrations/explore/IntegrationsExploreSection'

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    { render: props => <ExtensionsExploreSection {...props} /> },
    { render: props => <IntegrationsExploreSection {...props} /> },
    ...exploreSections,
]
