import * as GQL from '../../../shared/src/graphql/schema'
import * as React from 'react'
import { eventLogger } from '../tracking/eventLogger'
import { fetchLsifJobs, fetchLsifJobStatistics } from './backend'
import { Link } from '../../../shared/src/components/Link'
import { PageTitle } from '../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Subscription, Subject } from 'rxjs'
import {
    FilteredConnection,
    FilteredConnectionQueryArgs,
    FilteredConnectionFilter,
} from '../components/FilteredConnection'
import { Timestamp } from '../components/time/Timestamp'
import { switchMap, tap } from 'rxjs/operators'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

interface LsifJobNodeProps {
    node: GQL.ILSIFJob
}

interface LsifJobNodeState {}

class LsifJobNode extends React.PureComponent<LsifJobNodeProps, LsifJobNodeState> {
    public state: LsifJobNodeState = {}

    public render(): JSX.Element | null {
        return (
            <li className="lsif-job list-group-item">
                <div className="lsif-job__row lsif-job__main">
                    <div className="lsif-job__meta">
                        <div className="lsif-job__meta-root">
                            <Link to={`/site-admin/lsif-jobs/${this.props.node.id}`}>
                                {lsifJobDescription(this.props.node)}
                            </Link>
                        </div>
                    </div>

                    <small className="text-muted lsif-job__meta-timestamp">
                        <Link to={`/site-admin/lsif-jobs/${this.props.node.id}`}>
                            <Timestamp
                                noAbout={true}
                                date={
                                    this.props.node.finishedOn ||
                                    this.props.node.processedOn ||
                                    this.props.node.timestamp
                                }
                            />
                        </Link>
                    </small>
                </div>
            </li>
        )
    }
}

interface Props extends RouteComponentProps<any> {}

interface State {
    stats?: GQL.ILSIFJobStats
    loading: boolean
    error?: Error
}

class FilteredLsifJobsConnection extends FilteredConnection<{}, LsifJobNodeProps> {}

/**
 * A page displaying LSIF job activity.
 */
export class SiteAdminLsifJobsPage extends React.Component<Props, State> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'Processing',
            id: 'processing',
            tooltip: 'Show processing jobs only',
            args: { state: 'processing' },
        },
        {
            label: 'Errored',
            id: 'errored',
            tooltip: 'Show errored jobs only',
            args: { state: 'errored' },
        },
        {
            label: 'Completed',
            id: 'completed',
            tooltip: 'Show completed jobs only',
            args: { state: 'completed' },
        },
        {
            label: 'Queued',
            id: 'queued',
            tooltip: 'Show queued jobs only',
            args: { state: 'queued' },
        },
        {
            label: 'Scheduled',
            id: 'scheduled',
            tooltip: 'Show scheduled jobs only',
            args: { state: 'scheduled' },
        },
    ]

    public state: State = { loading: true }

    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminLsifJobPage')

        this.subscriptions.add(
            this.updates
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(() => fetchLsifJobStatistics())
                )
                .subscribe(
                    stats => this.setState({ stats, loading: false }),
                    error => this.setState({ error, loading: false })
                )
        )
        this.updates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-lsif-queues-page">
                <PageTitle title="LSIF Jobs - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">LSIF Jobs</h2>
                </div>

                {this.state.error && (
                    <div className="alert alert-danger">
                        Error getting LSIF job queue stats:
                        <br />
                        <code>{this.state.error.message}</code>
                    </div>
                )}
                {!this.state.stats && this.state.loading && <LoadingSpinner className="icon-inline" />}
                {!this.state.error && this.state.stats && (
                    <>
                        <div>
                            {this.state.stats.processingCount} processing, {this.state.stats.erroredCount} errored{' '}
                            {this.state.stats.completedCount} completed, {this.state.stats.queuedCount} queued, and{' '}
                            {this.state.stats.scheduledCount} scheduled
                        </div>
                    </>
                )}

                <FilteredLsifJobsConnection
                    noun="job"
                    pluralNoun="jobs"
                    queryConnection={this.queryJobs}
                    nodeComponent={LsifJobNode}
                    filters={SiteAdminLsifJobsPage.FILTERS}
                    history={this.props.history}
                    location={this.props.location}
                    listClassName="list-group list-group-flush"
                    cursorPaging={true}
                />
            </div>
        )
    }

    private queryJobs = (args: FilteredConnectionQueryArgs & { state?: string }) => {
        // Also refresh stats on each fetch
        this.updates.next()
        return fetchLsifJobs({ state: args.state || 'completed', ...args })
    }
}

/**
 * Construct a meaningful description from the job name and args.
 *
 * @param job The job instance.
 */
export function lsifJobDescription(job: GQL.ILSIFJob): JSX.Element {
    if (job.name === 'convert') {
        const { repository, commit, root } = job.args as {
            repository: string
            commit: string
            root: string
        }

        return (
            <span>
                [Convert] {repository} {root !== '' && root !== '/' && <>at {root}</>}{' '}
                <small>{commit.substring(0, 7)}</small>
            </span>
        )
    }

    const internalJobs: { [K: string]: string } = {
        'clean-old-jobs': 'Purge old job data from LSIF work queue',
        'clean-failed-jobs': 'Clean old failed job uploads from disk',
        'update-tips': 'Refresh current uploads',
    }

    if (internalJobs[job.name]) {
        return <span>[Internal] {internalJobs[job.name]}</span>
    }

    return <span>Unknown job type {job.name}</span>
}
