import React from 'react'
import { ExploreSectionDescriptor } from '../../../packages/webapp/src/explore/ExploreArea'
import { exploreSections } from '../../../packages/webapp/src/explore/exploreSections'
import { ExtensionsExploreSection } from '../extensions/explore/ExtensionsExploreSection'

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    { render: props => <ExtensionsExploreSection {...props} /> },
    ...exploreSections,
]
