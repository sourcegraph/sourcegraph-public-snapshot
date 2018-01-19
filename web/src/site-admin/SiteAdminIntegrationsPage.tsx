import Copy from '@sourcegraph/icons/lib/Copy'
import copy from 'copy-to-clipboard'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { browserExtensionMessageReceived } from '../tracking/analyticsUtils'
import { eventLogger } from '../tracking/eventLogger'

interface Props {}

interface State {
    browserExtensionInstalled: boolean
    copied: boolean
}

/**
 * A page displaying Sourcegraph Server integrations.
 */
export class SiteAdminIntegrationsPage extends React.Component<Props, State> {
    public state: State = {
        browserExtensionInstalled: false,
        copied: false,
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
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private sendMessage = () => {
        copy(window.context.appURL)
        this.setState({ copied: true })

        setTimeout(() => {
            this.setState({ copied: false })
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
                <div className="site-admin-integrations__section">
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
                <button className="btn btn-primary" onClick={this.sendMessage}>
                    <Copy className="site-admin-integrations__copy-icon icon-inline" />
                    {this.state.copied ? 'Copied' : 'Copy'}
                </button>
            </div>
        )
    }
}
