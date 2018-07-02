import * as React from 'react'
import { Observable, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import { CopyableText } from '../../components/CopyableText'
import { PageTitle } from '../../components/PageTitle'
import { IS_CHROME, IS_FIREFOX } from '../../marketing/util'
import { browserExtensionMessageReceived } from '../../tracking/analyticsUtils'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'

interface Props {}

interface State {
    browserExtensionInstalled: boolean
    openSearchUrlCopied: boolean
    info: { hasCodeIntelligence: boolean }
}

/**
 * A page displaying Sourcegraph integrations.
 */
export class UserAccountIntegrationsPage extends React.Component<Props, State> {
    public state: State = {
        browserExtensionInstalled: false,
        openSearchUrlCopied: false,
        info: {
            hasCodeIntelligence: false,
        },
    }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('UserAccountIntegrations')
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
                    {IS_CHROME && (
                        <a
                            className={`btn btn-primary mr-2 ${this.state.browserExtensionInstalled ? 'disabled' : ''}`}
                            target="_blank"
                            href="https://chrome.google.com/webstore/detail/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en"
                            onClick={this.installChromeExtension}
                        >
                            Install Chrome extension
                        </a>
                    )}
                    {IS_FIREFOX && (
                        <a
                            className={`btn btn-primary ${this.state.browserExtensionInstalled ? 'disabled' : ''}`}
                            target="_blank"
                            href="https://addons.mozilla.org/en-US/firefox/addon/sourcegraph/"
                            onClick={this.installFirefoxExtension}
                        >
                            Install Firefox add-on
                        </a>
                    )}
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
                        <strong>With the Chrome extension installed:</strong> Type <code>src</code> <kbd>SPACE</kbd> in
                        the address bar to search Sourcegraph.
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
                <CopyableText text={window.context.appURL} size={52} />
            </div>
        )
    }
}
