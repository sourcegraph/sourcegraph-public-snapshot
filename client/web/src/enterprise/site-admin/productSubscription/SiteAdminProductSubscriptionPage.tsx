import React, { useEffect } from 'react'

import { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'

import { ProductSubscriptionStatus } from './ProductSubscriptionStatus'

/**
 * Displays the product subscription information from the license key in site configuration.
 */
export const SiteAdminProductSubscriptionPage: React.FunctionComponent<TelemetryV2Props> = props => {
    useEffect(() => {
        props.telemetryRecorder.recordEvent('siteAdminProductSubscription', 'viewed')
        eventLogger.logViewEvent('SiteAdminProductSubscription')
    }, [props.telemetryRecorder])

    return (
        <div className="site-admin-product-subscription-page">
            <PageTitle title="Sourcegraph product subscription" />
            <ProductSubscriptionStatus showTrueUpStatus={true} />
        </div>
    )
}
