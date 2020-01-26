import {
    siteAdminOverviewComponents,
    SiteAdminOverviewComponent,
} from '../../../site-admin/overview/overviewComponents'
import { lazyComponent } from '../../../util/lazyComponent'

export const enterpriseSiteAdminOverviewComponents: readonly SiteAdminOverviewComponent[] = [
    {
        component: lazyComponent(
            () => import('../productSubscription/ProductSubscriptionStatus'),
            'ProductSubscriptionStatus'
        ),
        noCardClass: true,
        fullWidth: true,
    },
    ...siteAdminOverviewComponents,
]
