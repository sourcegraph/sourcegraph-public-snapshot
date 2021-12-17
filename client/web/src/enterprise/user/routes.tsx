import React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'

import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute, UserAreaRouteContext } from '../../user/area/UserArea'
import { lazyComponent } from '../../util/lazyComponent'
import { CreateBatchChangePageProps } from '../batches/create/CreateBatchChangePage'
import { CreateOrEditBatchChangePageProps } from '../batches/create/CreateOrEditBatchChangePage'
import { NamespaceBatchChangesAreaProps } from '../batches/global/GlobalBatchChangesArea'
import { SHOW_BUSINESS_FEATURES } from '../dotcom/productSubscriptions/features'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'

const NamespaceBatchChangesArea = lazyComponent<NamespaceBatchChangesAreaProps, 'NamespaceBatchChangesArea'>(
    () => import('../batches/global/GlobalBatchChangesArea'),
    'NamespaceBatchChangesArea'
)

const CreateOrEditBatchChangePage = lazyComponent<CreateOrEditBatchChangePageProps, 'CreateOrEditBatchChangePage'>(
    () => import('../batches/create/CreateOrEditBatchChangePage'),
    'CreateOrEditBatchChangePage'
)

const CreateBatchChangePage = lazyComponent<CreateBatchChangePageProps, 'CreateBatchChangePage'>(
    () => import('../batches/create/CreateBatchChangePage'),
    'CreateBatchChangePage'
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
            <CreateOrEditBatchChangePage
                {...props}
                initialNamespaceID={props.user.id}
                batchChangeName={match.params.batchChangeName}
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
