import React from 'react'
import { ExploreSectionDescriptor } from '../../explore/ExploreArea'
import { exploreSections } from '../../explore/exploreSections'
import { ExtensionsExploreSection } from '../extensions/explore/ExtensionsExploreSection'

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    { render: props => <ExtensionsExploreSection {...props} /> },
    ...exploreSections,
]
