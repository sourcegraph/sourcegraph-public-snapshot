import { ExploreSectionDescriptor } from '../../explore/ExploreArea'
import { exploreSections } from '../../explore/exploreSections'
import { asyncComponent } from '../../util/asyncComponent'

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    {
        render: asyncComponent(
            () => import('../extensions/explore/ExtensionsExploreSection'),
            'ExtensionsExploreSection'
        ),
    },
    ...exploreSections,
]
