import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import { PageTitle } from '../../../components/PageTitle'
import { eventLogger } from '../../../tracking/eventLogger'
import { ProductSubscriptionStatus } from './ProductSubscriptionStatus'

interface Props extends RouteComponentProps<{}> {}

/**
 * Displays the product subscription information from the license key in site configuration.
 */
export class SiteAdminProductSubscriptionPage extends React.Component<Props> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminProductSubscription')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const licenseKey = new URLSearchParams(this.props.location.search).get('licenseKey')
        return (
            <div className="site-admin-product-subscription-page">
                <PageTitle title="Sourcegraph product subscription" />
                {licenseKey && (
                    <div className="alert alert-success">
                        <strong>Your trial has started!</strong>
                        <p>
                            Your trial license key has been added to site configuration. You've unlocked more features.
                        </p>
                    </div>
                )}
                <ProductSubscriptionStatus showTrueUpStatus={true} />
            </div>
        )
    }
}
