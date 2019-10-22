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
                    <>
                        <p>id: {this.state.job.id}</p>
                        <p>name: {this.state.job.name}</p>
                        <p>args: {JSON.stringify(this.state.job.args)}</p>
                        <p>status: {this.state.job.status}</p>
                        <p>progress: {this.state.job.progress}</p>
                        <p>failedReason: {this.state.job.failedReason}</p>
                        <p>stacktrace: {this.state.job.stacktrace}</p>
                        <p>
                            timestamp:{' '}
                            {this.state.job.timestamp && (
                                <>
                                    {' '}
                                    <Timestamp date={this.state.job.timestamp} />{' '}
                                </>
                            )}
                        </p>
                        <p>
                            processedOn:{' '}
                            {this.state.job.processedOn && (
                                <>
                                    {' '}
                                    <Timestamp date={this.state.job.processedOn} />{' '}
                                </>
                            )}
                        </p>
                        <p>
                            finishedOn:{' '}
                            {this.state.job.finishedOn && (
                                <>
                                    {' '}
                                    <Timestamp date={this.state.job.finishedOn} />{' '}
                                </>
                            )}
                        </p>
                    </>
                )}
            </div>
        )
    }
}
