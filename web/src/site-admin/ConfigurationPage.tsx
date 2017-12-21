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
                    <div className="alert alert-danger site-admin-configuration-page__alert">
                        <p>
                            <strong>Configuration editing with JSON validation is enabled (beta)</strong>
                        </p>
                        <p>
                            Applying and saving configuration changes is not yet supported. Your changes will be lost
                            when you leave this page. You must manually copy the modified configuration and provide it
                            to the server (via <code>SOURCEGRAPH_CONFIG</code>).
                        </p>
                        <p>Use Cmd/Ctrl+Space for completion, and hover over JSON properties for documentation.</p>
                    </div>
                ) : (
                    <div className="alert alert-primary site-admin-configuration-page__alert">
                        <p>Configuration is read-only. Hover over JSON properties for documentation.</p>
                        <button className="btn btn-primary btn-sm" onClick={this.enableEditing}>
                            Enable editing (beta)
                        </button>
                    </div>
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
