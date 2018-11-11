import { siteAdminOverviewComponents } from '../../../web/src/site-admin/overviewComponents'
import { ProductSubscriptionStatus } from './productSubscription/ProductSubscriptionStatus'

export const enterpriseSiteAdminOverviewComponents: ReadonlyArray<React.ComponentType> = [
    ...siteAdminOverviewComponents,
    ProductSubscriptionStatus,
]
