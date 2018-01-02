import CheckmarkIcon from '@sourcegraph/icons/lib/Checkmark'
import DownloadIcon from '@sourcegraph/icons/lib/Download'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import { parse } from '@sqs/jsonc-parser/lib/main'
import formatDistance from 'date-fns/formatDistance'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { withLatestFrom } from 'rxjs/operators/withLatestFrom'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { SiteConfiguration } from '../schema/site.schema'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite, fetchSiteUpdateCheck } from './backend'
import { getTelemetryEnabled } from './SiteAdminTelemetryPage'

interface Props extends RouteComponentProps<any> {}

export interface State {
    channel?: string | null
    telemetryEnabled?: boolean
    version?: string
    updateCheck?: GQL.IUpdateCheck
    error?: string
}

/**
 * A page displaying information about available updates for the server.
 */
export class SiteAdminUpdatesPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminUpdates')

        this.subscriptions.add(
            fetchSite()
                .pipe(withLatestFrom(fetchSiteUpdateCheck()))
                .subscribe(
                    ([site, { version, updateCheck }]) =>
                        this.setState({
                            channel: getUpdateChannel(site.configuration.effectiveContents),
                            telemetryEnabled: getTelemetryEnabled(site.configuration.effectiveContents),
                            version,
                            updateCheck,
                            error: undefined,
                        }),
                    error => this.setState({ error: error.message })
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const autoUpdateCheckingEnabled = !!this.state.channel && this.state.telemetryEnabled
        return (
            <div className="site-admin-updates-page">
                <PageTitle title="Updates - Admin" />
                <h2>Updates</h2>
                {this.state.error && <p className="site-admin-updates-page__error">Error: {this.state.error}</p>}
                {this.state.updateCheck &&
                    (this.state.updateCheck.pending || this.state.updateCheck.checkedAt) && (
                        <div>
                            {this.state.updateCheck.pending && (
                                <div className="alert alert-primary">
                                    <LoaderIcon className="icon-inline" /> Checking for updates... (reload in a few
                                    seconds)
                                </div>
                            )}
                            {!this.state.updateCheck.errorMessage &&
                                (this.state.updateCheck.updateVersionAvailable ? (
                                    <div className="alert alert-success">
                                        <DownloadIcon className="icon-inline" /> Update available:{' '}
                                        <a href="https://about.sourcegraph.com">
                                            Sourcegraph Server {this.state.updateCheck.updateVersionAvailable}
                                        </a>
                                    </div>
                                ) : (
                                    <div className="alert alert-success">
                                        <CheckmarkIcon className="icon-inline" /> Sourcegraph Server is up to date.
                                    </div>
                                ))}
                            {this.state.updateCheck.errorMessage && (
                                <div className="alert alert-danger">
                                    Error checking for updates: {this.state.updateCheck.errorMessage}
                                </div>
                            )}
                        </div>
                    )}
                {!autoUpdateCheckingEnabled && (
                    <div className="alert alert-warning">Automatic update checking is disabled.</div>
                )}
                {this.state.updateCheck && (
                    <p className="site-admin-updates_page__info">
                        <small>
                            <strong>Current version:</strong> {this.state.version}
                        </small>
                        <br />
                        <small>
                            <strong>Last update check:</strong>{' '}
                            {this.state.updateCheck.checkedAt
                                ? formatDistance(this.state.updateCheck.checkedAt, new Date(), { addSuffix: true })
                                : 'never'}.
                        </small>
                        <br />
                        <small>
                            <strong>Automatic update checking:</strong> {this.state.channel ? 'on' : 'off'}.{' '}
                            <Link to="/site-admin/configuration">Configure</Link> <code>update.channel</code>{' '}
                            {!this.state.telemetryEnabled && (
                                <span>
                                    and <code>disableTelemetry</code>
                                </span>
                            )}{' '}
                            to {this.state.channel ? 'disable' : 'enable'}.
                        </small>
                    </p>
                )}
            </div>
        )
    }
}

function getUpdateChannel(cfgText: string): string | null {
    const o = parse(cfgText, [], { allowTrailingComma: true, disallowComments: false })
    return (o && (o as SiteConfiguration)['update.channel']) || null
}
