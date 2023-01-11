import React from 'react'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'

export const enterpriseSiteAdminOverviewComponents: readonly React.ComponentType<React.PropsWithChildren<any>>[] = [
    lazyComponent(() => import('../productSubscription/ProductSubscriptionStatus'), 'ProductSubscriptionStatus'),
]
