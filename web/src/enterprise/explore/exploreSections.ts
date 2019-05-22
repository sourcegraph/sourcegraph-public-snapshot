import { ExploreSectionDescriptor } from '../../explore/ExploreArea'
import { exploreSections } from '../../explore/exploreSections'
import { lazyComponent } from '../../util/lazyComponent'

export const enterpriseExploreSections: ReadonlyArray<ExploreSectionDescriptor> = [
    {
        render: lazyComponent(
            () => import('../extensions/explore/ExtensionsExploreSection'),
            'ExtensionsExploreSection'
        ),
    },
    ...exploreSections,
]
