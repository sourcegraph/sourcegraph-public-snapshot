import React, { useEffect } from 'react'

import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'

import { ProductSubscriptionStatus } from './ProductSubscriptionStatus'

/**
 * Displays the product subscription information from the license key in site configuration.
 */
export const SiteAdminProductSubscriptionPage: React.FunctionComponent = () => {
    useEffect(() => {
        window.context.telemetryRecorder?.recordEvent('siteAdminProductSubscription', 'viewed')
        eventLogger.logViewEvent('SiteAdminProductSubscription')
    }, [window.context.telemetryRecorder])

    return (
        <div className="site-admin-product-subscription-page">
            <PageTitle title="Sourcegraph product subscription" />
            <ProductSubscriptionStatus showTrueUpStatus={true} />
        </div>
    )
}
