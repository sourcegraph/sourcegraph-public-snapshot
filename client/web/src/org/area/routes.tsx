import { Navigate } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import type { EditBatchSpecPageProps } from '../../enterprise/batches/batch-spec/edit/EditBatchSpecPage'
import type { CreateBatchChangePageProps } from '../../enterprise/batches/create/CreateBatchChangePage'
import type { NamespaceBatchChangesAreaProps } from '../../enterprise/batches/global/GlobalBatchChangesArea'
import { namespaceAreaRoutes } from '../../namespaces/routes'

import type { OrgAreaRoute } from './OrgArea'

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

const OrgSettingsArea = lazyComponent(() => import('../settings/OrgSettingsArea'), 'OrgSettingsArea')

export const orgAreaRoutes: readonly OrgAreaRoute[] = [
    {
        path: 'getstarted',
        render: props => <Navigate to={`/organizations/${props.org.name}/settings/members`} replace={true} />,
    },
    {
        path: 'settings/*',
        render: props => (
            <OrgSettingsArea
                {...props}
                routes={props.orgSettingsAreaRoutes}
                sideBarItems={props.orgSettingsSideBarItems}
            />
        ),
    },
    ...namespaceAreaRoutes,

    // Redirect from /organizations/:orgname -> /organizations/:orgname/settings/profile.
    {
        path: '',
        render: () => <Navigate to="./settings/profile" />,
    },
    // Redirect from previous /organizations/:orgname/account -> /organizations/:orgname/settings/profile.
    {
        path: 'account',
        render: () => <Navigate to="../settings/profile" />,
    },

    {
        path: 'batch-changes/create',
        render: props => <CreateBatchChangePage headingElement="h1" {...props} initialNamespaceID={props.org.id} />,
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
        render: props => <NamespaceBatchChangesArea {...props} namespaceID={props.org.id} />,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
    },
]
