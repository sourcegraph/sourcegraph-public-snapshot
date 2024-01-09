import { Navigate } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'
import type { EditBatchSpecPageProps } from '../../enterprise/batches/batch-spec/edit/EditBatchSpecPage'
import type { CreateBatchChangePageProps } from '../../enterprise/batches/create/CreateBatchChangePage'
import type { NamespaceBatchChangesAreaProps } from '../../enterprise/batches/global/GlobalBatchChangesArea'
import { SHOW_BUSINESS_FEATURES } from '../../enterprise/dotcom/productSubscriptions/features'
import { namespaceAreaRoutes } from '../../namespaces/routes'

import type { UserAreaRoute } from './UserArea'

const NamespaceBatchChangesArea = lazyComponent<NamespaceBatchChangesAreaProps, 'NamespaceBatchChangesArea'>(
    () => import('../../enterprise/batches/global/GlobalBatchChangesArea'),
    'NamespaceBatchChangesArea'
)

const ExecuteBatchSpecPage = lazyComponent(
    () => import('../../enterprise/batches/batch-spec/execute/ExecuteBatchSpecPage'),
    'ExecuteBatchSpecPage'
)

const CreateBatchChangePage = lazyComponent<CreateBatchChangePageProps, 'CreateBatchChangePage'>(
    () => import('../../enterprise/batches/create/CreateBatchChangePage'),
    'CreateBatchChangePage'
)

const EditBatchSpecPage = lazyComponent<EditBatchSpecPageProps, 'EditBatchSpecPage'>(
    () => import('../../enterprise/batches/batch-spec/edit/EditBatchSpecPage'),
    'EditBatchSpecPage'
)

const UserSettingsArea = lazyComponent(() => import('../settings/UserSettingsArea'), 'UserSettingsArea')
const UserProfile = lazyComponent(() => import('../profile/UserProfile'), 'UserProfile')

export const userAreaRoutes: readonly UserAreaRoute[] = [
    {
        path: 'settings/*',
        render: props => (
            <UserSettingsArea
                {...props}
                routes={props.userSettingsAreaRoutes}
                sideBarItems={props.userSettingsSideBarItems}
                telemetryRecorder={props.platformContext.telemetryRecorder}
            />
        ),
    },
    {
        path: 'profile',
        render: props => <UserProfile user={props.user} />,
    },
    ...namespaceAreaRoutes,

    // Redirect from /users/:username -> /users/:username/profile.
    {
        path: '',
        render: () => <Navigate to="profile" replace={true} />,
    },
    // Redirect from previous /users/:username/account -> /users/:username/profile.
    {
        path: 'account',
        render: () => <Navigate to="../profile" replace={true} />,
    },

    // Redirect from previous /users/:username/subscriptions -> /users/:username/settings/subscriptions.
    {
        path: 'subscriptions/:page?',
        render: () => (
            <RedirectRoute
                getRedirectURL={({ params }) => `../settings/subscriptions${params.page ? `/${params.page}` : ''}`}
            />
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: 'batch-changes/create',
        render: props => <CreateBatchChangePage headingElement="h1" {...props} initialNamespaceID={props.user.id} />,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        fullPage: true,
    },
    {
        path: 'batch-changes/:batchChangeName/edit',
        render: props => <EditBatchSpecPage {...props} />,
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: 'batch-changes/:batchChangeName/executions/:batchSpecID/*',
        render: props => <ExecuteBatchSpecPage {...props} />,
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: 'batch-changes/*',
        render: props => <NamespaceBatchChangesArea {...props} namespaceID={props.user.id} />,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
    },
]
