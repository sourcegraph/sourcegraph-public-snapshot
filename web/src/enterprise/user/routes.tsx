import React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute, UserAreaRouteContext } from '../../user/area/UserArea'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'
import { UserCampaignListPageProps } from '../campaigns/list/CampaignListPage'
import { lazyComponent } from '../../util/lazyComponent'
import { CampaignApplyPageProps } from '../campaigns/apply/CampaignApplyPage'
import { CampaignDetailsProps } from '../campaigns/detail/CampaignDetails'

const UserCampaignListPage = lazyComponent<UserCampaignListPageProps, 'UserCampaignListPage'>(
    () => import('../campaigns/list/CampaignListPage'),
    'UserCampaignListPage'
)
const CampaignApplyPage = lazyComponent<CampaignApplyPageProps, 'CampaignApplyPage'>(
    () => import('../campaigns/apply/CampaignApplyPage'),
    'CampaignApplyPage'
)
const CampaignDetails = lazyComponent<CampaignDetailsProps, 'CampaignDetails'>(
    () => import('../campaigns/detail/CampaignDetails'),
    'CampaignDetails'
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
        path: '/campaigns/apply/:specID',
        render: ({ match, ...props }: UserAreaRouteContext & RouteComponentProps<{ specID: string }>) => (
            <div className="web-content">
                <CampaignApplyPage {...props} specID={match.params.specID} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) =>
            !isSourcegraphDotCom && window.context.experimentalFeatures?.automation === 'enabled',
    },
    {
        path: '/campaigns/:campaignID',
        render: ({ match, ...props }: UserAreaRouteContext & RouteComponentProps<{ campaignID: string }>) => (
            <div className="web-content">
                <CampaignDetails {...props} campaignID={match.params.campaignID} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) =>
            !isSourcegraphDotCom && window.context.experimentalFeatures?.automation === 'enabled',
    },
    {
        path: '/campaigns',
        render: props => (
            <div className="web-content">
                <UserCampaignListPage {...props} userID={props.user.id} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) =>
            !isSourcegraphDotCom && window.context.experimentalFeatures?.automation === 'enabled',
    },
]
