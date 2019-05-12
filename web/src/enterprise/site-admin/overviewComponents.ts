import React from 'react'
import { siteAdminOverviewComponents } from '../../site-admin/overviewComponents'
import { asyncComponent } from '../../util/asyncComponent'

export const enterpriseSiteAdminOverviewComponents: ReadonlyArray<React.ComponentType<any>> = [
    ...siteAdminOverviewComponents,
    asyncComponent(() => import('./productSubscription/ProductSubscriptionStatus'), 'ProductSubscriptionStatus'),
]
