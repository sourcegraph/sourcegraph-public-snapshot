import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

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
        render: props => <SavedSearchListPage {...props} telemetryRecorder={props.platformContext.telemetryRecorder} />,
        condition: ({ license }) => license.isCodeSearchEnabled,
    },
    {
        path: 'searches/add',
        render: props => (
            <SavedSearchCreateForm {...props} telemetryRecorder={props.platformContext.telemetryRecorder} />
        ),
        condition: ({ license }) => license.isCodeSearchEnabled,
    },
    {
        path: 'searches/:id',
        render: props => (
            <SavedSearchUpdateForm {...props} telemetryRecorder={props.platformContext.telemetryRecorder} />
        ),
        condition: ({ license }) => license.isCodeSearchEnabled,
    },
]
