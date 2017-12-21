import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { MonacoSettingsEditor } from '../settings/MonacoSettingsEditor'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite } from './backend'

interface Props extends RouteComponentProps<any> {}

interface State {
    site?: GQL.ISite
    editing?: boolean
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
                {this.state.editing ? (
                    <p className="alert alert-danger">
                        <strong>Configuration editing with JSON validation is enabled (beta)</strong>
                        <br />
                        <br />
                        Applying and saving configuration changes is not yet supported. Your changes will be lost when
                        you leave this page. You must manually copy the modified configuration and provide it to the
                        server (via <code>SOURCEGRAPH_CONFIG</code>).
                        <br />
                        <br />
                        Use Cmd/Ctrl+Space for completion, and hover over JSON properties for documentation.
                    </p>
                ) : (
                    <p className="alert alert-primary">
                        Configuration is read-only. Hover over JSON properties for documentation. <br />
                        <br />
                        <a className="site-admin-configuration-page__enable-editing" onClick={this.enableEditing}>
                            Enable editing (beta)
                        </a>
                    </p>
                )}
                {this.state.site &&
                    this.state.site.latestSettings && (
                        <MonacoSettingsEditor
                            className="site-admin-configuration-page__config"
                            value={this.state.site.configuration}
                            jsonSchema="https://sourcegraph.com/v1/site.schema.json#"
                            readOnly={!this.state.editing}
                            height={700}
                        />
                    )}
            </div>
        )
    }

    private enableEditing = () => this.setState({ editing: true })
}
