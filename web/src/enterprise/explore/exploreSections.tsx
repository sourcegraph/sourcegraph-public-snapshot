import { ExploreSectionDescriptor } from '../../explore/ExploreArea'
import { exploreSections } from '../../explore/exploreSections'
import { asyncComponent } from '../../util/asyncComponent'

const ExtensionsExploreSection = asyncComponent(
    () => import('../extensions/explore/ExtensionsExploreSection'),
    'ExtensionsExploreSection'
)

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    { render: props => <ExtensionsExploreSection {...props} /> },
    ...exploreSections,
]
