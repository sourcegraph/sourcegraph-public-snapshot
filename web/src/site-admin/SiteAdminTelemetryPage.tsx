import Loader from '@sourcegraph/icons/lib/Loader'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite } from './backend'
import { getTelemetryEnabled } from './configHelpers'

class TelemetrySample extends React.PureComponent<
    { children: string; telemetryEnabled: boolean },
    { expanded: boolean }
> {
    public state = { expanded: false }

    public render(): JSX.Element | null {
        const [firstLine, ...otherLines] = this.props.children.split('\n')
        return (
            <div className={`telemetry-sample ${this.state.expanded ? 'telemetry-sample-expanded' : ''}`}>
                <pre
                    className={`telemetry-sample__body ${this.state.expanded ? 'telemetry-sample__body-expanded' : ''}`}
                >
                    <small>
                        {firstLine} {!this.props.telemetryEnabled && '(not sent)'}
                        {'\n'}
                        {this.state.expanded && otherLines.join('\n')}
                    </small>
                </pre>
                <div className="telemetry-sample__btn-container">
                    <button className="btn btn-secondary btn-sm telemetry-sample__btn" onClick={this.toggle}>
                        {this.state.expanded ? 'Hide' : 'Show'}
                    </button>
                </div>
            </div>
        )
    }

    private toggle = () => {
        this.setState(state => ({ expanded: !state.expanded }))
    }
}

interface Props extends RouteComponentProps<any> {}

export interface State {
    telemetryEnabled?: boolean
    telemetrySamples?: string[]
    error?: string
}

/**
 * A page displaying information about telemetry for the site.
 */
export class SiteAdminTelemetryPage extends React.Component<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminTelemetry')

        this.subscriptions.add(
            fetchSite({ telemetrySamples: true }).subscribe(
                site =>
                    this.setState({
                        telemetryEnabled: getTelemetryEnabled(site.configuration.effectiveContents),
                        telemetrySamples: site.telemetrySamples,
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
            <div className="site-admin-telemetry-page">
                <PageTitle title="Telemetry - Admin" />
                <h2>Telemetry</h2>
                {this.state.error && (
                    <p className="site-admin-telemetry-page__error">Error: {upperFirst(this.state.error)}</p>
                )}
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
                {!this.state.telemetrySamples && <Loader className="icon-inline" />}
                {this.state.telemetrySamples &&
                    (this.state.telemetrySamples.length > 0 ? (
                        <div>
                            <h3>Samples</h3>
                            <p>
                                Inspect recent telemetry samples to see what{' '}
                                {this.state.telemetryEnabled ? 'is' : 'would be'} sent.{' '}
                                {!this.state.telemetryEnabled &&
                                    'Because telemetry is disabled, this information is not being sent.'}
                            </p>
                            {this.state.telemetrySamples.map((sample, i) => (
                                <TelemetrySample key={i} telemetryEnabled={!!this.state.telemetryEnabled}>
                                    {sample}
                                </TelemetrySample>
                            ))}
                        </div>
                    ) : (
                        <p>No telemetry samples have been collected.</p>
                    ))}
            </div>
        )
    }
}
