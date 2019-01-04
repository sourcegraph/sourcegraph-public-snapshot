import React from 'react'
import { siteAdminOverviewComponents } from '../../site-admin/overviewComponents'
const ProductSubscriptionStatus = React.lazy(async () => ({
    default: (await import('./productSubscription/ProductSubscriptionStatus')).ProductSubscriptionStatus,
}))

export const enterpriseSiteAdminOverviewComponents: ReadonlyArray<React.ComponentType<any>> = [
    ...siteAdminOverviewComponents,
    ProductSubscriptionStatus,
]
