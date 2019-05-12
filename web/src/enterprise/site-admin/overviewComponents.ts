import React from 'react'
import { siteAdminOverviewComponents } from '../../site-admin/overviewComponents'
import { asyncComponent } from '../../util/asyncComponent'

const ProductSubscriptionStatus = asyncComponent(
    () => import('./productSubscription/ProductSubscriptionStatus'),
    'ProductSubscriptionStatus'
)

export const enterpriseSiteAdminOverviewComponents: ReadonlyArray<React.ComponentType<any>> = [
    ...siteAdminOverviewComponents,
    ProductSubscriptionStatus,
]
