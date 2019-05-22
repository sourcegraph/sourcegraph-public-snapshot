import React from 'react'
import { siteAdminOverviewComponents } from '../../site-admin/overviewComponents'
import { lazyComponent } from '../../util/lazyComponent'

export const enterpriseSiteAdminOverviewComponents: ReadonlyArray<React.ComponentType<any>> = [
    ...siteAdminOverviewComponents,
    lazyComponent(() => import('./productSubscription/ProductSubscriptionStatus'), 'ProductSubscriptionStatus'),
]
