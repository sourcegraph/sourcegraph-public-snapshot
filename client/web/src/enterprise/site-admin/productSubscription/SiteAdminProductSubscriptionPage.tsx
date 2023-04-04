import React, { useEffect } from 'react'

import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'

import { ProductSubscriptionStatus } from './ProductSubscriptionStatus'

/**
 * Displays the product subscription information from the license key in site configuration.
 */
export const SiteAdminProductSubscriptionPage: React.FunctionComponent = () => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminProductSubscription'), [])

    return (
        <div className="site-admin-product-subscription-page">
            <PageTitle title="Sourcegraph product subscription" />
            <ProductSubscriptionStatus showTrueUpStatus={true} />
        </div>
    )
}
