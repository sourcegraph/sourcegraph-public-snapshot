import { parse } from '@sqs/jsonc-parser/lib/main'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { SiteConfiguration } from '../schema/site.schema'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite } from './backend'

interface Props extends RouteComponentProps<any> {}

export interface State {
    telemetryEnabled?: boolean
    error?: string
}

/**
 * A page displaying information about telemetry for the site.
 */
export class TelemetryPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminTelemetry')

        this.subscriptions.add(
            fetchSite().subscribe(
                site =>
                    this.setState({
                        telemetryEnabled: getTelemetryEnabled(site.configuration.effectiveContents),
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
        return (
            <div className="telemetry-page">
                <PageTitle title="Telemetry - Admin" />
                <h2>Telemetry</h2>
                {this.state.error && <p className="telemetry-page__error">Error: {this.state.error}</p>}
                {typeof this.state.telemetryEnabled === 'boolean' && (
                    <div>
                        <p>
                            {this.state.telemetryEnabled
                                ? 'Telemetry is enabled. Product usage and performance data will be sent to Sourcegraph to help improve the product.'
                                : 'Telemetry is disabled. No product usage or performance data will be sent to Sourcegraph.'}
                        </p>
                        <p>No code, file names, repository names, search queries, or settings values are ever sent.</p>
                        <p>
                            Set <code>disableTelemetry</code> to <code>{this.state.telemetryEnabled.toString()}</code>{' '}
                            in <Link to="/site-admin/configuration">site configuration</Link> to{' '}
                            {this.state.telemetryEnabled ? 'disable' : 'enable'} telemetry.
                        </p>
                    </div>
                )}
            </div>
        )
    }
}

/** Parses out the 'disableTelemetry' key from the JSON site config and returns the inverse. */
function getTelemetryEnabled(text: string): boolean {
    const o = parse(text, [], { allowTrailingComma: true, disallowComments: false })
    return o && !(o as SiteConfiguration).disableTelemetry
}
