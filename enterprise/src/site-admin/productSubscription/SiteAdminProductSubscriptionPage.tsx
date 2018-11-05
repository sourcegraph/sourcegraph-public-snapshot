import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import { PageTitle } from '../../../../packages/webapp/src/components/PageTitle'
import { eventLogger } from '../../../../packages/webapp/src/tracking/eventLogger'
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
        return (
            <div className="site-admin-product-subscription-page">
                <PageTitle title="Sourcegraph product subscription" />
                <ProductSubscriptionStatus showTrueUpStatus={true} />
            </div>
        )
    }
}
