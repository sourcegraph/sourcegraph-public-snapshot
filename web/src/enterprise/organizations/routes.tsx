import React from 'react'
import { OrgAreaRoute } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'
import { lazyComponent } from '../../util/lazyComponent'
import { OrgCampaignListPageProps } from '../campaigns/global/list/GlobalCampaignListPage'

const OrgCampaignListPage = lazyComponent<OrgCampaignListPageProps, 'OrgCampaignListPage'>(
    () => import('../campaigns/global/list/GlobalCampaignListPage'),
    'OrgCampaignListPage'
)

export const enterpriseOrganizationAreaRoutes: readonly OrgAreaRoute[] = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
    {
        path: '/campaigns',
        render: props => <OrgCampaignListPage {...props} orgID={props.org.id} />,
    },
]
