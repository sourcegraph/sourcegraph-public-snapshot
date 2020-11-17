import React from 'react'
import { OrgAreaRoute } from '../../org/area/OrgArea'
import { orgAreaRoutes } from '../../org/area/routes'
import { enterpriseNamespaceAreaRoutes } from '../namespaces/routes'
import { lazyComponent } from '../../util/lazyComponent'
import { NamespaceCampaignsAreaProps } from '../campaigns/global/GlobalCampaignsArea'

const NamespaceCampaignsArea = lazyComponent<NamespaceCampaignsAreaProps, 'NamespaceCampaignsArea'>(
    () => import('../campaigns/global/GlobalCampaignsArea'),
    'NamespaceCampaignsArea'
)

export const enterpriseOrganizationAreaRoutes: readonly OrgAreaRoute[] = [
    ...orgAreaRoutes,
    ...enterpriseNamespaceAreaRoutes,
    {
        path: '/campaigns',
        render: props => <NamespaceCampaignsArea {...props} namespaceID={props.org.id} />,
        condition: props => !props.isSourcegraphDotCom && window.context.campaignsEnabled,
    },
]
