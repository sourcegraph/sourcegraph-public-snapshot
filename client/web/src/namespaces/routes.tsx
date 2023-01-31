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
        v5: {
            path: '/searches',
            exact: true,
            render: props => <SavedSearchListPage {...props} />,
        },
        v6: {
            path: '/searches',
            render: props => <SavedSearchListPage {...props} />,
        },
    },
    {
        v5: {
            path: '/searches/add',
            render: props => <SavedSearchCreateForm {...props} />,
        },
        v6: {
            path: '/searches/add',
            render: props => <SavedSearchCreateForm {...props} />,
        },
    },
    {
        v5: {
            path: '/searches/:id',
            render: props => <SavedSearchUpdateForm {...props} />,
        },
        v6: {
            path: '/searches/:id',
            render: props => <SavedSearchUpdateForm {...props} />,
        },
    },
]
