import Copy from '@sourcegraph/icons/lib/Copy'
import copy from 'copy-to-clipboard'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { IS_CHROME, IS_FIREFOX } from '../../marketing/util'
import { browserExtensionMessageReceived } from '../../tracking/analyticsUtils'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'

interface Props {}

interface State {
    browserExtensionInstalled: boolean
    remoteUrlCopied: boolean
    openSearchUrlCopied: boolean
    info: { hasCodeIntelligence: boolean }
}

/**
 * A page displaying Sourcegraph Server integrations.
 */
export class UserSettingsIntegrationsPage extends React.Component<Props, State> {
    public state: State = {
        browserExtensionInstalled: false,
        remoteUrlCopied: false,
        openSearchUrlCopied: false,
        info: {
            hasCodeIntelligence: false,
        },
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserSettingsIntegrations')
        this.subscriptions.add(
            browserExtensionMessageReceived.subscribe(isInstalled => {
                this.setState(() => ({ browserExtensionInstalled: isInstalled }))
            })
        )

        this.subscriptions.add(this.fetchOverview().subscribe(info => this.setState(() => ({ info }))))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private copyRemoteUrl = () => {
        eventLogger.log('CopySourcegraphURLClicked')
        copy(window.context.appURL)
        this.setState({ remoteUrlCopied: true })

        setTimeout(() => {
            this.setState({ remoteUrlCopied: false })
        }, 2000)
    }

    private installFirefoxExtension = () => {
        eventLogger.log('BrowserExtInstallClicked', { marketing: { browser: 'Firefox' } })
    }

    private installChromeExtension = () => {
        eventLogger.log('BrowserExtInstallClicked', { marketing: { browser: 'Chrome' } })
    }

    private fetchOverview(): Observable<{ hasCodeIntelligence: boolean }> {
        return queryGraphQL(gql`
            query Overview {
                site {
                    hasCodeIntelligence
                }
            }
        `).pipe(
            map(({ data, errors }) => {
                if (!data || !data.site) {
                    throw createAggregateError(errors)
                }
                return {
                    hasCodeIntelligence: data.site.hasCodeIntelligence,
                }
            })
        )
    }

    private enableCodeIntelligence = () => {
        eventLogger.log('EnableCodeIntelligenceClicked')
    }

    public render(): JSX.Element | null {
        return (
            <div>
                <PageTitle title="Integrations" />
                <h2>Integrations</h2>

                <h3>Browser extension</h3>
                <p>
                    <a
                        className={`btn btn-primary mr-2 ${
                            this.state.browserExtensionInstalled && IS_CHROME ? 'disabled' : ''
                        }`}
                        target="_blank"
                        href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en"
                        onClick={this.installChromeExtension}
                    >
                        Install Chrome extension
                    </a>
                    <a
                        className={`btn btn-primary ${
                            this.state.browserExtensionInstalled && IS_FIREFOX ? 'disabled' : ''
                        }`}
                        target="_blank"
                        href="https://addons.mozilla.org/en-US/firefox/addon/sourcegraph-addon-for-github/"
                        onClick={this.installFirefoxExtension}
                    >
                        Install Firefox add-on
                    </a>
                </p>
                {!this.state.info.hasCodeIntelligence && (
                    <div className="alert alert-warning">
                        <a
                            href="https://about.sourcegraph.com/docs/code-intelligence/install?utm_source=integrations"
                            target="_blank"
                            onClick={this.enableCodeIntelligence}
                        >
                            Enable code intelligence
                        </a>{' '}
                        to use the full functionality of the browser extensions.
                    </div>
                )}

                <h3 className="mt-2">Browser address bar search shortcut</h3>
                <ul className="list ml-3">
                    <li>
                        <strong>With the Chrome extension installed:</strong> Type <kbd>src &lt;SPACE&gt;</kbd> in the
                        address bar to search Sourcegraph.
                    </li>
                    <li>
                        <strong>Other browsers:</strong> Add Sourcegraph as a search engine using the URL{' '}
                        <code>{window.context.appURL}/search?q=%s</code> (assuming <code>%s</code> is the query
                        placeholder).
                    </li>
                </ul>
                <p>
                    Need further customization? In Chrome, visit <strong>{'chrome://settings/searchEngines'}</strong> or{' '}
                    <a
                        target="_blank"
                        href="https://support.google.com/chrome/answer/95426?hl=en&co=GENIE.Platform=Desktop"
                    >
                        documentation
                    </a>. In Firefox, visit <strong>{'about:preferences#search'}</strong> or{' '}
                    <a target="_blank" href="https://support.mozilla.org/en-US/kb/add-or-remove-search-engine-firefox">
                        documentation
                    </a>.
                </p>

                <h3 className="mt-2">Editor extensions and other integrations</h3>
                <p>Use the following Sourcegraph URL to connect other extensions to this Sourcegraph instance.</p>
                <div>
                    <input
                        id="remote-url-input"
                        className="form-control"
                        value={window.context.appURL}
                        readOnly={true}
                    />
                    <button className="btn btn-primary mt-2" onClick={this.copyRemoteUrl}>
                        <Copy className="icon-inline" /> {this.state.remoteUrlCopied ? 'Copied' : 'Copy'}
                    </button>
                </div>
            </div>
        )
    }
}
