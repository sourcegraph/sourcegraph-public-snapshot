import React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute, UserAreaRouteContext } from '../../user/area/UserArea'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'
import { UserCampaignListPageProps } from '../campaigns/global/list/GlobalCampaignListPage'
import { lazyComponent } from '../../util/lazyComponent'

const UserCampaignListPage = lazyComponent<UserCampaignListPageProps, 'UserCampaignListPage'>(
    () => import('../campaigns/global/list/GlobalCampaignListPage'),
    'UserCampaignListPage'
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
    },
    {
        path: '/campaigns',
        render: props => <UserCampaignListPage {...props} userID={props.user.id} />,
        condition: ({ isSourcegraphDotCom }) =>
            !isSourcegraphDotCom && window.context.experimentalFeatures?.automation === 'enabled',
    },
]
