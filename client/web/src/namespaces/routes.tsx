import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { NamespaceAreaRoute } from './NamespaceArea'

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
        path: '/searches',
        render: props => <SavedSearchListPage {...props} />,
        // TODO: Remove once RR6 migration is finished. For now these work on both RR5 and RR6.
        exact: true,
    },
    {
        path: '/searches/add',
        render: props => <SavedSearchCreateForm {...props} />,
        // TODO: Remove once RR6 migration is finished. For now these work on both RR5 and RR6.
        exact: true,
    },
    {
        path: '/searches/:id',
        render: props => <SavedSearchUpdateForm {...props} />,
        // TODO: Remove once RR6 migration is finished. For now these work on both RR5 and RR6.
        exact: true,
    },
]
