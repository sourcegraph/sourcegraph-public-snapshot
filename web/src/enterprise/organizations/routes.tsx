import React from 'react'
import { OrgAreaRoute, OrgAreaPageProps } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'
import { lazyComponent } from '../../util/lazyComponent'
import { OrgCampaignListPageProps } from '../campaigns/global/list/GlobalCampaignListPage'
import { CampaignApplyPageProps } from '../campaigns/apply/CampaignApplyPage'
import { RouteComponentProps } from 'react-router'

const OrgCampaignListPage = lazyComponent<OrgCampaignListPageProps, 'OrgCampaignListPage'>(
    () => import('../campaigns/global/list/GlobalCampaignListPage'),
    'OrgCampaignListPage'
)
const CampaignApplyPage = lazyComponent<CampaignApplyPageProps, 'CampaignApplyPage'>(
    () => import('../campaigns/apply/CampaignApplyPage'),
    'CampaignApplyPage'
)

export const enterpriseOrganizationAreaRoutes: readonly OrgAreaRoute[] = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
    {
        path: '/campaigns/apply/:specID',
        render: ({ match, ...props }: OrgAreaPageProps & RouteComponentProps<{ specID: string }>) => (
            <CampaignApplyPage {...props} specID={match.params.specID} />
        ),
        condition: ({ isSourcegraphDotCom }) =>
            !isSourcegraphDotCom && window.context.experimentalFeatures?.automation === 'enabled',
    },
    {
        path: '/campaigns',
        render: props => <OrgCampaignListPage {...props} orgID={props.org.id} />,
        condition: ({ isSourcegraphDotCom }) =>
            !isSourcegraphDotCom && window.context.experimentalFeatures?.automation === 'enabled',
    },
]
