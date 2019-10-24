import * as GQL from '../../../shared/src/graphql/schema'
import * as LSIF from '../../../shared/src/lsif/description'
import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'
import { fetchLsifJob } from './backend'
import { PageTitle } from '../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'
import { upperFirst } from 'lodash'
import { Timestamp } from '../components/time/Timestamp'

interface Props extends RouteComponentProps<{ id: string }> {}

interface State {
    job?: GQL.ILsifJob | null
    error?: Error
}

/**
 * A page displaying metadata about an LSIF job.
 */
export class SiteAdminLsifJobPage extends React.Component<Props, State> {
    constructor(props: Props) {
        super(props)
    }

    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminLsifJob')

        this.subscriptions.add(
            fetchLsifJob(this.props.match.params.id).subscribe(
                job => this.setState({ job }),
                error => this.setState({ error })
            )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-lsif-jobs-page">
                <PageTitle title="LSIF Jobs - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">{this.state.job && LSIF.lsifJobDescription(this.state.job)}</h2>
                </div>

                {this.state.error && <p className="alert alert-danger">{upperFirst(this.state.error.message)}</p>}
                {this.state.job && (
                    <table className="table repo-settings-index-page__stats">
                        <tbody>
                            <tr>
                                <th>ID</th>
                                <td>{this.state.job.id}</td>
                            </tr>
                            <tr>
                                <th>Name</th>
                                <td>{this.state.job.name}</td>
                            </tr>

                            <tr>
                                <th>Args</th>
                                <td>
                                    {this.state.job.args && (
                                        <table>
                                            {Object.entries(this.state.job.args)
                                                .sort(([a], [b]) => a.localeCompare(b))
                                                .map(([key, value]) => (
                                                    <>
                                                        <tr>
                                                            <th>{key}</th>
                                                            <td>{value}</td>
                                                        </tr>
                                                    </>
                                                ))}
                                        </table>
                                    )}
                                </td>
                            </tr>
                            <tr>
                                <th>Status</th>
                                <td>{this.state.job.status}</td>
                            </tr>
                            <tr>
                                <th>Progress</th>
                                <td>{this.state.job.progress}</td>
                            </tr>
                            {this.state.job.failedReason && (
                                <>
                                    <tr>
                                        <th>Failed Reason</th>
                                        <td>{this.state.job.failedReason}</td>
                                    </tr>

                                    <tr>
                                        <th>Stacktrace</th>
                                        <td>
                                            <pre>{this.state.job.stacktrace}</pre>
                                        </td>
                                    </tr>
                                </>
                            )}
                            {this.state.job.timestamp && (
                                <tr>
                                    <th>Timestamp</th>
                                    <td>
                                        <Timestamp date={this.state.job.timestamp} />
                                    </td>
                                </tr>
                            )}
                            {this.state.job.processedOn && (
                                <tr>
                                    <th>Processed On</th>
                                    <td>
                                        <Timestamp date={this.state.job.processedOn} />
                                    </td>
                                </tr>
                            )}
                            {this.state.job.finishedOn && (
                                <tr>
                                    <th>Finished On</th>
                                    <td>
                                        <Timestamp date={this.state.job.finishedOn} />
                                    </td>
                                </tr>
                            )}
                        </tbody>
                    </table>
                )}
            </div>
        )
    }
}
