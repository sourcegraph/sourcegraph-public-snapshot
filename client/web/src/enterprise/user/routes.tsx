import { Redirect, RouteComponentProps } from 'react-router'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute, UserAreaRouteContext } from '../../user/area/UserArea'
import { EditBatchSpecPageProps } from '../batches/batch-spec/edit/EditBatchSpecPage'
import { CreateBatchChangePageProps } from '../batches/create/CreateBatchChangePage'
import { NamespaceBatchChangesAreaProps } from '../batches/global/GlobalBatchChangesArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'

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

    // Redirect from previous /users/:username/subscriptions -> /users/:username/settings/subscriptions.
    {
        path: '/subscriptions/:page*',
        render: (props: UserAreaRouteContext & RouteComponentProps<{ page: string }>) => (
            <Redirect
                to={`${props.url}/settings/subscriptions${
                    props.match.params.page ? `/${props.match.params.page}` : ''
                }`}
            />
        ),
        condition: () => SHOW_BUSINESS_FEATURES,
    },
    {
        path: '/batch-changes/create',
        render: props => <CreateBatchChangePage headingElement="h1" {...props} initialNamespaceID={props.user.id} />,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
        fullPage: true,
    },
    {
        path: '/batch-changes/:batchChangeName/edit',
        render: ({ match, ...props }: UserAreaRouteContext & RouteComponentProps<{ batchChangeName: string }>) => (
            <EditBatchSpecPage
                {...props}
                batchChange={{ name: match.params.batchChangeName, namespace: props.user.id }}
            />
        ),
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: '/batch-changes/:batchChangeName/executions/:batchSpecID',
        render: ({
            match,
            ...props
        }: UserAreaRouteContext & RouteComponentProps<{ batchChangeName: string; batchSpecID: string }>) => (
            <ExecuteBatchSpecPage
                {...props}
                batchSpecID={match.params.batchSpecID}
                batchChange={{ name: match.params.batchChangeName, namespace: props.user.id }}
                match={match}
            />
        ),
        condition: ({ batchChangesEnabled, batchChangesExecutionEnabled }) =>
            batchChangesEnabled && batchChangesExecutionEnabled,
        fullPage: true,
    },
    {
        path: '/batch-changes',
        render: props => <NamespaceBatchChangesArea {...props} namespaceID={props.user.id} />,
        condition: ({ batchChangesEnabled }) => batchChangesEnabled,
    },
]
