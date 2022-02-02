import React from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

import { siteAdminOverviewComponents } from '../../../site-admin/overview/overviewComponents'

export const enterpriseSiteAdminOverviewComponents: readonly React.ComponentType<any>[] = [
    ...siteAdminOverviewComponents,
    lazyComponent(() => import('../productSubscription/ProductSubscriptionStatus'), 'ProductSubscriptionStatus'),
]
