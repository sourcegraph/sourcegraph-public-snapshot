import React, { useEffect } from 'react'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { ProductSubscriptionStatus } from './ProductSubscriptionStatus'
import { RouteComponentProps } from 'react-router'

/**
 * Displays the product subscription information from the license key in site configuration.
 */
export const SiteAdminProductSubscriptionPage: React.FunctionComponent<RouteComponentProps> = ({ location }) => {
    useEffect(() => eventLogger.logViewEvent('SiteAdminProductSubscription'), [])

    const licenseKey = new URLSearchParams(location.search).get('licenseKey')

    return (
        <div className="site-admin-product-subscription-page">
            <PageTitle title="Sourcegraph product subscription" />
            {licenseKey && (
                <div className="alert alert-success">
                    <strong>Your trial has started!</strong>
                    <p>Your trial license key has been added to site configuration. You've unlocked more features.</p>
                </div>
            )}
            <ProductSubscriptionStatus showTrueUpStatus={true} />
        </div>
    )
}
