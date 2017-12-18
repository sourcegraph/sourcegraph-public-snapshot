import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

interface Props extends RouteComponentProps<any> {}

/**
 * A page displaying an overview of site admin information.
 */
export class OverviewPage extends React.Component<Props> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminOverview')
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-overview-page">
                <PageTitle title="Overview - Admin" />
                <p>Welcome to the site administration area!</p>
            </div>
        )
    }
}
