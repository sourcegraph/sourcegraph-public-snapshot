import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite } from './backend'

interface Props extends RouteComponentProps<any> {}

interface State {
    site?: GQL.ISite
}

/**
 * A page displaying the site configuration.
 */
export class ConfigurationPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminConfiguration')

        this.subscriptions.add(fetchSite().subscribe(site => this.setState({ site })))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-configuration-page">
                <PageTitle title="Configuration - Admin" />
                <h2>Configuration</h2>
                <p>
                    Configuration is specified as a string of JSON passed in the <code>SOURCEGRAPH_CONFIG</code>{' '}
                    environment variable. See{' '}
                    <a href="https://about.sourcegraph.com/docs/server/">Sourcegraph configuration documentation</a>.
                </p>
                {this.state.site &&
                    this.state.site.latestSettings && (
                        <pre className="site-admin-configuration-page__config">
                            <code>{this.state.site.configuration}</code>
                        </pre>
                    )}
            </div>
        )
    }
}
