import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'
import { userAreaRoutes } from '../../user/area/routes'
import type { UserAreaRoute } from '../../user/area/UserArea'
import type { EditBatchSpecPageProps } from '../batches/batch-spec/edit/EditBatchSpecPage'
import type { CreateBatchChangePageProps } from '../batches/create/CreateBatchChangePage'
import type { NamespaceBatchChangesAreaProps } from '../batches/global/GlobalBatchChangesArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'

const AppSettingsArea = lazyComponent(() => import('../app/settings/AppSettingsArea'), 'AppSettingsArea')

const NamespaceBatchChangesArea = lazyComponent<NamespaceBatchChangesAreaProps, 'NamespaceBatchChangesArea'>(
    () => import('../batches/global/GlobalBatchChangesArea'),
    'NamespaceBatchChangesArea'
)

const ExecuteBatchSpecPage = lazyComponent(
    () => import('../batches/batch-spec/execute/ExecuteBatchSpecPage'),
    'ExecuteBatchSpecPage'
)

const CreateBatchChangePage = lazyComponent<CreateBatchChangePageProps, 'CreateBatchChangePage'>(
    () => import('../batches/create/CreateBatchChangePage'),
    'CreateBatchChangePage'
)

const EditBatchSpecPage = lazyComponent<EditBatchSpecPageProps, 'EditBatchSpecPage'>(
    () => import('../batches/batch-spec/edit/EditBatchSpecPage'),
    'EditBatchSpecPage'
)

export const enterpriseUserAreaRoutes: readonly UserAreaRoute[] = [
    ...userAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,

    // Cody app specific route (cody/app settings page)
    // This route won't be available for any non-app deploy types.
    // See userAreaHeaderNavItems in client/web/src/enterprise/user/navitems.ts
    // for more context on user settings page.
    {
        path: 'app-settings/*',
        render: props => (
            <AppSettingsArea telemetryService={props.telemetryService} telemetryRecorder={props.telemetryRecorder} />
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
