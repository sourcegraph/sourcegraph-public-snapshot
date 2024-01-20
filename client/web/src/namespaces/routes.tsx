import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { isCodyOnlyLicense } from '../util/license'

import type { NamespaceAreaRoute } from './NamespaceArea'

const SavedSearchListPage = lazyComponent(() => import('../savedSearches/SavedSearchListPage'), 'SavedSearchListPage')

const SavedSearchCreateForm = lazyComponent(
    () => import('../savedSearches/SavedSearchCreateForm'),
    'SavedSearchCreateForm'
)
const SavedSearchUpdateForm = lazyComponent(
    () => import('../savedSearches/SavedSearchUpdateForm'),
    'SavedSearchUpdateForm'
)

export const namespaceAreaRoutes: readonly NamespaceAreaRoute[] = [
    {
        path: 'searches',
        render: props => <SavedSearchListPage {...props} />,
        condition: () => !isCodyOnlyLicense(),
    },
    {
        path: 'searches/add',
        render: props => <SavedSearchCreateForm {...props} />,
        condition: () => !isCodyOnlyLicense(),
    },
    {
        path: 'searches/:id',
        render: props => <SavedSearchUpdateForm {...props} />,
        condition: () => !isCodyOnlyLicense(),
    },
]
