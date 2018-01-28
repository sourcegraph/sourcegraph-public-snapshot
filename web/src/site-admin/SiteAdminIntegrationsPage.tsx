import Copy from '@sourcegraph/icons/lib/Copy'
import copy from 'copy-to-clipboard'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { browserExtensionMessageReceived } from '../tracking/analyticsUtils'
import { eventLogger } from '../tracking/eventLogger'
import { fetchOpenSearchSettings } from './backend'

interface Props {}

interface State {
    browserExtensionInstalled: boolean
    remoteUrlCopied: boolean
    openSearchSettings?: GQL.IOpenSearchSettings | null
}

/**
 * A page displaying Sourcegraph Server integrations.
 */
export class SiteAdminIntegrationsPage extends React.Component<Props, State> {
    public state: State = {
        browserExtensionInstalled: false,
        remoteUrlCopied: false,
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminIntegrations')
        this.subscriptions.add(
            browserExtensionMessageReceived.subscribe(isInstalled => {
                this.setState({
                    browserExtensionInstalled: isInstalled,
                })
            })
        )

        this.subscriptions.add(
            fetchOpenSearchSettings().subscribe(openSearchSettings => this.setState({ openSearchSettings }))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private copyRemoteUrl = () => {
        copy(window.context.appURL)
        this.setState({ remoteUrlCopied: true })

        setTimeout(() => {
            this.setState({ remoteUrlCopied: false })
        }, 2000)
    }

    private installExtension = () => {
        window.open(
            'https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en',
            '__blank'
        )
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-detail-list">
                <PageTitle title="Integrations - Admin" />
                <h2 className="site-admin-integrations__header">Integrations</h2>
                <h3 className="site-admin-integrations__integration-name">Browser extension</h3>
                <p>
                    The Sourcegraph Firefox add-on and Chrome extension add code search and code intelligence on GitHub,
                    including pull request diffs. By default, the extension will add code intelligence and code search
                    to public repositories. The extension can be configured to work on private code by connecting it to
                    a Sourcegraph Server instance.
                </p>
                <div className="site-admin-page__actions">
                    <button
                        className="btn btn-primary"
                        onClick={this.installExtension}
                        disabled={this.state.browserExtensionInstalled}
                    >
                        Install browser extension
                    </button>
                </div>
                <div className="site-admin-integrations__sub-section">
                    <h4 className="input-label">Sourcegraph Server URL</h4>
                    To connect the browser extension to your Sourcegraph Server instance, click the extension icon.
                    Then, fill in the Sourcegraph URL field with your Sourcegraph Server URL and hit save.
                </div>
                <div className="site-admin-integrations__form">
                    <input
                        id="remote-url-input"
                        className="site-admin-integrations__input form-control"
                        value={window.context.appURL}
                        readOnly={true}
                    />
                </div>
                <button className="btn btn-primary" onClick={this.copyRemoteUrl}>
                    <Copy className="site-admin-integrations__copy-icon icon-inline" />
                    {this.state.remoteUrlCopied ? 'Copied' : 'Copy'}
                </button>
                <div className="site-admin-integrations__section">
                    <h3 className="site-admin-integrations__integration-name">
                        Code search from Google Chrome address bar
                    </h3>
                    <p>
                        Get easy access to Sourcegraph code search over your most commonly-used repositories from your
                        browser address bar. Type your keyword or Sourcegraph Server URL into the address bar and then
                        press <kbd>tab</kbd> to use Sourcegraph code search.
                    </p>
                    <b>Sourcegraph Server Search URL</b>
                    <div className="site-admin-integrations__form">
                        <input
                            id="remote-url-input"
                            className="site-admin-integrations__input form-control"
                            value={
                                this.state.openSearchSettings
                                    ? this.state.openSearchSettings.searchURL
                                    : `${window.context.appURL}/search?q={searchTerms}`
                            }
                            readOnly={true}
                        />
                    </div>
                    <div>
                        To edit your address bar search URL click the <b>Edit</b> button below, then select the{' '}
                        <b>Add address bar search</b>.
                    </div>
                    <div className="site-admin-integrations__sub-section">
                        <Link to="/site-admin/configuration">
                            <button className="btn btn-primary">Edit</button>
                        </Link>
                    </div>
                    <div className="site-admin-integrations__section">
                        <b>Additional Configuration</b>
                        <p>
                            To customize the address bar keyword, navigate to{' '}
                            <a href="chrome://settings/searchEngines">{'chrome://settings/searchEngines'}</a>.
                        </p>
                    </div>
                </div>
            </div>
        )
    }
}
