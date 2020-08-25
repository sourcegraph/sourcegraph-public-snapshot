import React from 'react'
import { OrgAreaRoute, OrgAreaPageProps } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'
import { lazyComponent } from '../../util/lazyComponent'
import { OrgCampaignListPageProps } from '../campaigns/list/CampaignListPage'
import { CampaignApplyPageProps } from '../campaigns/apply/CampaignApplyPage'
import { RouteComponentProps } from 'react-router'
import { CampaignDetailsProps } from '../campaigns/detail/CampaignDetails'
import { CampaignClosePageProps } from '../campaigns/close/CampaignClosePage'
import { CreateCampaignPageProps } from '../campaigns/create/CreateCampaignPage'

const OrgCampaignListPage = lazyComponent<OrgCampaignListPageProps, 'OrgCampaignListPage'>(
    () => import('../campaigns/list/CampaignListPage'),
    'OrgCampaignListPage'
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

export const enterpriseOrganizationAreaRoutes: readonly OrgAreaRoute[] = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
    {
        path: '/campaigns/apply/:specID',
        render: ({ match, ...props }: OrgAreaPageProps & RouteComponentProps<{ specID: string }>) => (
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
        render: ({ match, ...props }: OrgAreaPageProps & RouteComponentProps<{ campaignID: string }>) => (
            <div className="web-content">
                <CampaignClosePage {...props} campaignID={match.params.campaignID} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
    {
        path: '/campaigns/:campaignID',
        render: ({ match, ...props }: OrgAreaPageProps & RouteComponentProps<{ campaignID: string }>) => (
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
                <OrgCampaignListPage {...props} orgID={props.org.id} />
            </div>
        ),
        condition: ({ isSourcegraphDotCom }) => !isSourcegraphDotCom && window.context.campaignsEnabled,
    },
]
