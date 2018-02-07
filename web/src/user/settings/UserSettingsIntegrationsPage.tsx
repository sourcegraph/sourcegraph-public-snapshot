import Copy from '@sourcegraph/icons/lib/Copy'
import copy from 'copy-to-clipboard'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { IS_CHROME, IS_FIREFOX } from '../../marketing/util'
import { browserExtensionInstalled } from '../../tracking/analyticsUtils'
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
            browserExtensionInstalled.subscribe(isInstalled => {
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

    private copyOpenSearchUrl = () => {
        eventLogger.log('CopyOpenSearchURLClicked')
        copy(`${window.context.appURL}/search?q=%s`)
        this.setState({ openSearchUrlCopied: true })

        setTimeout(() => {
            this.setState({ openSearchUrlCopied: false })
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
        window.open('https://about.sourcegraph.com/pricing', 'blank')
    }

    private contactUsClicked = () => {
        eventLogger.log('ContactUsClicked')
        window.open('https://about.sourcegraph.com/contact/sales', 'blank')
    }

    public render(): JSX.Element | null {
        return (
            <div className="user-integrations__list">
                <PageTitle title="Integrations" />
                <h2 className="user-integrations__header">Integrations</h2>
                <h3 className="user-integrations__integration-name">Browser extension</h3>
                <p>
                    The Sourcegraph extension adds code search and code intelligence on GitHub, including pull request
                    diffs.
                </p>
                <div>
                    <div className="user-integrations__action-items">
                        <a
                            className={`btn btn-primary ${
                                this.state.browserExtensionInstalled && IS_CHROME ? 'disabled' : ''
                            }`}
                            target="_blank"
                            href="https://chrome.google.com/webstore/detail/sourcegraph-for-github/dgjhfomjieaadpoljlnidmbgkdffpack?hl=en"
                            onClick={this.installChromeExtension}
                        >
                            Install Chrome extension
                        </a>
                    </div>
                    <div className="user-integrations__action-items">
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
                    </div>
                </div>
                <table className="table-hover user-integrations__table">
                    <tbody>
                        <tr className="user-integrations__table-row">
                            <td className="user-integrations__table-item">Code search</td>
                            <td>
                                <div className="user-integrations__btn-container">
                                    <button className={`btn btn-secondary btn-sm`} disabled={true}>
                                        Enabled
                                    </button>
                                </div>
                            </td>
                        </tr>
                        <tr className="user-integrations__table-row">
                            <td className="user-integrations__table-item">Code intelligence</td>
                            <td>
                                <div className="user-integrations__btn-container">
                                    <button
                                        className={`btn btn-${
                                            this.state.info.hasCodeIntelligence ? 'secondary' : 'primary'
                                        } btn-sm`}
                                        onClick={this.enableCodeIntelligence}
                                        disabled={this.state.info.hasCodeIntelligence}
                                    >
                                        {this.state.info.hasCodeIntelligence ? 'Enabled' : 'Enable'}
                                    </button>
                                </div>
                            </td>
                        </tr>
                        <tr className="user-integrations__table-row">
                            <td className="user-integrations__table-item">GitHub</td>
                            <td>
                                <div className="user-integrations__btn-container">
                                    <button className={`btn btn-secondary btn-sm`} disabled={true}>
                                        Enabled
                                    </button>
                                </div>
                            </td>
                        </tr>
                        <tr className="user-integrations__table-row">
                            <td className="user-integrations__table-item">GitHub Enterprise</td>
                            <td>
                                <div className="user-integrations__btn-container">
                                    <button className={`btn btn-secondary btn-sm`} disabled={true}>
                                        Enabled
                                    </button>
                                </div>
                            </td>
                        </tr>
                        <tr className="user-integrations__table-row">
                            <td className="user-integrations__table-item">Other code hosts</td>
                            <td>
                                <div className="user-integrations__btn-container">
                                    <button className={`btn btn-primary btn-sm`} onClick={this.contactUsClicked}>
                                        Contact us
                                    </button>
                                </div>
                            </td>
                        </tr>
                    </tbody>
                </table>
                <div className="user-integrations__sub-section">
                    <h4 className="input-label">Sourcegraph Server URL</h4>
                    <p>
                        To connect the browser extension to your Sourcegraph Server instance, click the extension icon
                        and click <b>Connect</b>.
                    </p>
                    <p>
                        You can change your Sourcegraph URL at any time by clicking the extension icon and filling in
                        the Sourcegraph URL field with your Sourcegraph Server URL and clicking save.
                    </p>
                </div>
                <div className="user-integrations__form">
                    <input
                        id="remote-url-input"
                        className="user-integrations__input form-control"
                        value={window.context.appURL}
                        readOnly={true}
                    />
                </div>
                <button className="btn btn-primary" onClick={this.copyRemoteUrl}>
                    <Copy className="user-integrations__copy-icon icon-inline" />
                    {this.state.remoteUrlCopied ? 'Copied' : 'Copy'}
                </button>
                <div className="user-integrations__divider" />
                <div>
                    <h3 className="user-integrations__integration-name">Code search from address bar</h3>
                    <p>
                        Get easy access to Sourcegraph code search over your most commonly-used repositories from your
                        browser address bar by configuring a custom search engine.
                    </p>
                    <h4>Sourcegraph search URL</h4>
                    <div className="user-integrations__form">
                        <input
                            id="remote-url-input"
                            className="user-integrations__input form-control"
                            value={`${window.context.appURL}/search?q=%s`}
                            readOnly={true}
                        />
                        <button className="btn btn-primary" onClick={this.copyOpenSearchUrl}>
                            <Copy className="user-integrations__copy-icon icon-inline" />
                            {this.state.openSearchUrlCopied ? 'Copied' : 'Copy'}
                        </button>
                    </div>
                </div>
                <div className="user-integrations__sub-section">
                    <h4>Code search from Google Chrome address bar</h4>
                    Go to <b>{'chrome://settings/searchEngines'}</b> to add a custom search engine or
                    <a
                        target="_blank"
                        href="https://support.google.com/chrome/answer/95426?hl=en&co=GENIE.Platform=Desktop"
                    >
                        {' '}
                        read this Chrome page{' '}
                    </a>
                    for more information.
                </div>
                <div className="user-integrations__sub-section">
                    <h4>Add Firefox search engine</h4>
                    Go to <b>{'about:preferences#search'}</b> to add a one-click search engine or check out
                    <a target="_blank" href="https://support.mozilla.org/en-US/kb/add-or-remove-search-engine-firefox">
                        {' '}
                        this Firefox page{' '}
                    </a>
                    for additional documentation.
                </div>
            </div>
        )
    }
}
