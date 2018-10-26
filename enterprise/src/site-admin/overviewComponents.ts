import { siteAdminOverviewComponents } from '@sourcegraph/webapp/dist/site-admin/overviewComponents'
import { ProductSubscriptionStatus } from './productSubscription/ProductSubscriptionStatus'

export const enterpriseSiteAdminOverviewComponents: ReadonlyArray<React.ComponentType> = [
    ...siteAdminOverviewComponents,
    ProductSubscriptionStatus,
]
