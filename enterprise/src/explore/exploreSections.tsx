import React from 'react'
import { ExploreSectionDescriptor } from '../../../web/src/explore/ExploreArea'
import { exploreSections } from '../../../web/src/explore/exploreSections'
import { ExtensionsExploreSection } from '../extensions/explore/ExtensionsExploreSection'

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    { render: props => <ExtensionsExploreSection {...props} /> },
    ...exploreSections,
]
