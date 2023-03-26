import { isDefined } from '@sourcegraph/common'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { RedirectRoute } from '../../components/RedirectRoute'
import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute } from '../../user/area/UserArea'
import { EditBatchSpecPageProps } from '../batches/batch-spec/edit/EditBatchSpecPage'
import { CreateBatchChangePageProps } from '../batches/create/CreateBatchChangePage'
import { NamespaceBatchChangesAreaProps } from '../batches/global/GlobalBatchChangesArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'

const NamespaceBatchChangesArea = !process.env.DISABLE_BATCH_CHANGES
    ? lazyComponent<NamespaceBatchChangesAreaProps, 'NamespaceBatchChangesArea'>(
          () => import('../batches/global/GlobalBatchChangesArea'),
          'NamespaceBatchChangesArea'
      )
    : null

const ExecuteBatchSpecPage = !process.env.DISABLE_BATCH_CHANGES
    ? lazyComponent(() => import('../batches/batch-spec/execute/ExecuteBatchSpecPage'), 'ExecuteBatchSpecPage')
    : null

const CreateBatchChangePage = !process.env.DISABLE_BATCH_CHANGES
    ? lazyComponent<CreateBatchChangePageProps, 'CreateBatchChangePage'>(
          () => import('../batches/create/CreateBatchChangePage'),
          'CreateBatchChangePage'
      )
    : null

const EditBatchSpecPage = !process.env.DISABLE_BATCH_CHANGES
    ? lazyComponent<EditBatchSpecPageProps, 'EditBatchSpecPage'>(
          () => import('../batches/batch-spec/edit/EditBatchSpecPage'),
          'EditBatchSpecPage'
      )
    : null

export const enterpriseUserAreaRoutes: readonly UserAreaRoute[] = [
    ...userAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,

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
    CreateBatchChangePage &&
        ({
            path: 'batch-changes/create',
            render: props => (
                <CreateBatchChangePage headingElement="h1" {...props} initialNamespaceID={props.user.id} />
            ),
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
            fullPage: true,
        } satisfies UserAreaRoute),
    EditBatchSpecPage &&
        ({
            path: 'batch-changes/:batchChangeName/edit',
            render: props => <EditBatchSpecPage {...props} />,
            condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
                batchChangesEnabled && batchChangesExecutionEnabled,
            fullPage: true,
        } satisfies UserAreaRoute),
    ExecuteBatchSpecPage &&
        ({
            path: 'batch-changes/:batchChangeName/executions/:batchSpecID/*',
            render: props => <ExecuteBatchSpecPage {...props} />,
            condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
                batchChangesEnabled && batchChangesExecutionEnabled,
            fullPage: true,
        } satisfies UserAreaRoute),
    NamespaceBatchChangesArea &&
        ({
            path: 'batch-changes/*',
            render: props => <NamespaceBatchChangesArea {...props} namespaceID={props.user.id} />,
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        } satisfies UserAreaRoute),
].filter(isDefined)
