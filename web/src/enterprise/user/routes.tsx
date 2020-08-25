import React from 'react'
import { Redirect, RouteComponentProps } from 'react-router'
import { userAreaRoutes } from '../../user/area/routes'
import { UserAreaRoute, UserAreaRouteContext } from '../../user/area/UserArea'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'
import { UserCampaignListPageProps } from '../campaigns/list/CampaignListPage'
import { lazyComponent } from '../../util/lazyComponent'
import { CampaignApplyPageProps } from '../campaigns/apply/CampaignApplyPage'
import { CampaignDetailsProps } from '../campaigns/detail/CampaignDetails'
import { CampaignClosePageProps } from '../campaigns/close/CampaignClosePage'
import { CreateCampaignPageProps } from '../campaigns/create/CreateCampaignPage'

const UserCampaignListPage = lazyComponent<UserCampaignListPageProps, 'UserCampaignListPage'>(
    () => import('../campaigns/list/CampaignListPage'),
    'UserCampaignListPage'
)
const CampaignApplyPage = lazyComponent<CampaignApplyPageProps, 'CampaignApplyPage'>(
    () => import('../campaigns/apply/CampaignApplyPage'),
    'CampaignApplyPage'
)
const CreateCampaignPage = lazyComponent<CreateCampaignPageProps, 'CreateCampaignPage'>(
    () => import('../campaigns/create/CreateCampaignPage'),
    'CreateCampaignPage'
)
const CampaignDetails = lazyComponent<CampaignDetailsProps, 'CampaignDetails'>(
    () => import('../campaigns/detail/CampaignDetails'),
    'CampaignDetails'
)
const CampaignClosePage = lazyComponent<CampaignClosePageProps, 'CampaignClosePage'>(
    () => import('../campaigns/close/CampaignClosePage'),
    'CampaignClosePage'
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
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
    {
        path: '/campaigns/create',
        render: props => (
            <div className="web-content">
                <CreateCampaignPage {...props} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
    {
        path: '/campaigns/:campaignID/close',
        render: ({ match, ...props }: UserAreaRouteContext & RouteComponentProps<{ campaignID: string }>) => (
            <div className="web-content">
                <CampaignClosePage {...props} campaignID={match.params.campaignID} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
    {
        path: '/campaigns/:campaignID',
        render: ({ match, ...props }: UserAreaRouteContext & RouteComponentProps<{ campaignID: string }>) => (
            <div className="web-content">
                <CampaignDetails {...props} campaignID={match.params.campaignID} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
    {
        path: '/campaigns',
        render: props => (
            <div className="web-content">
                <UserCampaignListPage {...props} userID={props.user.id} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
]
