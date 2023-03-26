import { isDefined } from '@sourcegraph/common'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { OrgAreaRoute } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { EditBatchSpecPageProps } from '../batches/batch-spec/edit/EditBatchSpecPage'
import { CreateBatchChangePageProps } from '../batches/create/CreateBatchChangePage'
import { NamespaceBatchChangesAreaProps } from '../batches/global/GlobalBatchChangesArea'
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

export const enterpriseOrganizationAreaRoutes: readonly OrgAreaRoute[] = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
    CreateBatchChangePage &&
        ({
            path: 'batch-changes/create',
            render: props => <CreateBatchChangePage headingElement="h1" {...props} initialNamespaceID={props.org.id} />,
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
            fullPage: true,
        } satisfies OrgAreaRoute),
    EditBatchSpecPage &&
        ({
            path: 'batch-changes/:batchChangeName/edit',
            render: props => <EditBatchSpecPage {...props} />,
            condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
                batchChangesEnabled && batchChangesExecutionEnabled,
            fullPage: true,
        } satisfies OrgAreaRoute),
    ExecuteBatchSpecPage &&
        ({
            path: 'batch-changes/:batchChangeName/executions/:batchSpecID/*',
            render: props => <ExecuteBatchSpecPage {...props} />,
            condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
                batchChangesEnabled && batchChangesExecutionEnabled,
            fullPage: true,
        } satisfies OrgAreaRoute),
    NamespaceBatchChangesArea &&
        ({
            path: 'batch-changes/*',
            render: props => <NamespaceBatchChangesArea {...props} namespaceID={props.org.id} />,
            condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        } satisfies OrgAreaRoute),
].filter(isDefined)
