import { Navigate } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'
import type { EditBatchSpecPageProps } from '../../enterprise/batches/batch-spec/edit/EditBatchSpecPage'
import type { CreateBatchChangePageProps } from '../../enterprise/batches/create/CreateBatchChangePage'
import type { NamespaceBatchChangesAreaProps } from '../../enterprise/batches/global/GlobalBatchChangesArea'
import { SHOW_BUSINESS_FEATURES } from '../../enterprise/dotcom/productSubscriptions/features'
import { namespaceAreaRoutes } from '../../namespaces/routes'

import type { UserAreaRoute } from './UserArea'

const AppSettingsArea = lazyComponent(() => import('../../enterprise/app/settings/AppSettingsArea'), 'AppSettingsArea')

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
        render: props => <UserProfile user={props.user} telemetryRecorder={props.platformContext.telemetryRecorder} />,
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

    // Cody app specific route (cody/app settings page)
    // This route won't be available for any non-app deploy types.
    // See userAreaHeaderNavItems in client/web/src/enterprise/user/navitems.ts
    // for more context on user settings page.
    {
        path: 'app-settings/*',
        render: props => (
            <AppSettingsArea
                telemetryService={props.telemetryService}
                telemetryRecorder={props.platformContext.telemetryRecorder}
            />
        ),
        condition: context => context.isCodyApp,
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
        render: props => (
            <CreateBatchChangePage
                headingElement="h1"
                {...props}
                initialNamespaceID={props.user.id}
                telemetryRecorder={props.platformContext.telemetryRecorder}
            />
        ),
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        fullPage: true,
    },
    {
        path: 'batch-changes/:batchChangeName/edit',
        render: props => <EditBatchSpecPage {...props} telemetryRecorder={props.platformContext.telemetryRecorder} />,
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: 'batch-changes/:batchChangeName/executions/:batchSpecID/*',
        render: props => (
            <ExecuteBatchSpecPage {...props} telemetryRecorder={props.platformContext.telemetryRecorder} />
        ),
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: 'batch-changes/*',
        render: props => (
            <NamespaceBatchChangesArea
                {...props}
                namespaceID={props.user.id}
                telemetryRecorder={props.platformContext.telemetryRecorder}
            />
        ),
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
    },
]
