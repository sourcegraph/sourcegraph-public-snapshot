import React from 'react'
import { ExploreSectionDescriptor } from '../../explore/ExploreArea'
import { exploreSections } from '../../explore/exploreSections'
const ExtensionsExploreSection = React.lazy(async () => ({
    default: (await import('../extensions/explore/ExtensionsExploreSection')).ExtensionsExploreSection,
}))

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    { render: props => <ExtensionsExploreSection {...props} /> },
    ...exploreSections,
]
