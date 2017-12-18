import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'

interface Props extends RouteComponentProps<any> {}

/**
 * A page displaying the site configuration.
 */
export class ConfigurationPage extends React.Component<Props> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminConfiguration')
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
                    <a href="https://about.sourcegraph.com/docs/server/config/">
                        Sourcegraph configuration documentation
                    </a>.
                </p>
            </div>
        )
    }
}
