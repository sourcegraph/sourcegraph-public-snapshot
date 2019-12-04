import * as GQL from '../../../../shared/src/graphql/schema'
import React, { FunctionComponent } from 'react'
import { ErrorLike, asError } from '../../../../shared/src/util/errors'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifJobs, fetchLsifJobStatistics } from './backend'
import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { Link } from '../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription, Observable } from 'rxjs'
import { switchMap, catchError } from 'rxjs/operators'
import { Timestamp } from '../../components/time/Timestamp'
import { Toggle } from '../../../../shared/src/components/Toggle'
import {
    FilteredConnection,
    FilteredConnectionQueryArgs,
    FilteredConnectionFilter,
} from '../../components/FilteredConnection'
import { ErrorAlert } from '../../components/alerts'

interface ToggleComponentProps {
    hideInternal: boolean
    onToggle: (enabled: boolean) => void
}

const ToggleComponent: FunctionComponent<ToggleComponentProps> = ({ hideInternal, onToggle }) => (
    <div className="lsif-jobs-internal-toggle">
        <label className="lsif-jobs-internal-toggle__label" title="Hide internal jobs">
            <Toggle value={hideInternal} onToggle={onToggle} title="Hide internal jobs" />
            <small>
                <div className="lsif-jobs-internal-toggle__label-hint">Hide internal jobs</div>
            </small>
        </label>
    </div>
)

interface LsifJobNodeProps {
    node: GQL.ILSIFJob
}

const LsifJobNode: FunctionComponent<LsifJobNodeProps> = ({ node }) => (
    <li className="lsif-job list-group-item">
        <div className="lsif-job__row lsif-job__main">
            <div className="lsif-job__meta">
                <div className="lsif-job__meta-root">
                    <Link to={`/site-admin/lsif-jobs/${node.id}`}>{lsifJobDescription(node)}</Link>
                </div>
            </div>

            <small className="text-muted lsif-job__meta-timestamp">
                <Timestamp noAbout={true} date={node.completedOrErroredAt || node.startedAt || node.queuedAt} />
            </small>
        </div>
    </li>
)

interface Props extends RouteComponentProps<any> {}

interface State {
    statsOrError: GQL.ILSIFJobStats | ErrorLike | null
    hideInternal: boolean
}

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

    public state: State = {
        statsOrError: null,
        hideInternal: true,
    }

    /** Emits when internal jobs is switched on/off from view. */
    private toggles = new Subject<void>()
    /** Emits when stats should be refreshed. */
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminLsifJobs')

        this.subscriptions.add(
            this.updates
                // Do not set statsOrError as null here to indicate loading
                // so that the stats don't disappear during navigation, which
                // causes some weird jitter.
                .pipe(switchMap(() => fetchLsifJobStatistics().pipe(catchError(err => [asError(err)]))))
                .subscribe(stats => this.setState({ statsOrError: stats }))
        )
        this.updates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-lsif-jobs-page">
                <PageTitle title="LSIF jobs - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">LSIF jobs</h2>
                </div>
                {!this.state.statsOrError ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.statsOrError) ? (
                    <div className="alert alert-danger">
                        <ErrorAlert error={this.state.statsOrError} />
                    </div>
                ) : (
                    <div className="mb-3">
                        {this.state.statsOrError.processingCount} processing, {this.state.statsOrError.erroredCount},{' '}
                        errored {this.state.statsOrError.completedCount} completed,{' '}
                        {this.state.statsOrError.queuedCount} queued, and {this.state.statsOrError.scheduledCount}{' '}
                        scheduled
                    </div>
                )}

                <FilteredConnection<{}, LsifJobNodeProps>
                    noun="job"
                    pluralNoun="jobs"
                    updates={this.toggles}
                    queryConnection={this.queryJobs}
                    nodeComponent={LsifJobNode}
                    filters={SiteAdminLsifJobsPage.FILTERS}
                    additionalFilterElement={
                        <ToggleComponent hideInternal={this.state.hideInternal} onToggle={this.onToggle} />
                    }
                    history={this.props.history}
                    location={this.props.location}
                    listClassName="list-group list-group-flush mt-3"
                    cursorPaging={true}
                />
            </div>
        )
    }

    private onToggle = (hideInternal: boolean): void => {
        this.setState({ hideInternal }, () => this.toggles.next())
    }

    private queryJobs = (
        args: FilteredConnectionQueryArgs & { state?: GQL.LSIFJobState }
    ): Observable<GQL.ILSIFJobConnection> => {
        if (this.state.hideInternal) {
            // Simulate internal job filtering by searching for the only
            // job type that isn't an internal job.
            args.query = `convert ${args.query}`
        }

        // Also refresh stats on each fetch
        this.updates.next()

        // Do the actual query
        return fetchLsifJobs({
            // For typechecker: ensure state is not undefined
            state: args.state || GQL.LSIFJobState.COMPLETED,
            ...args,
        })
    }
}

/**
 * Construct a meaningful description from the job name and args.
 *
 * @param job The job instance.
 */
function lsifJobDescription(job: GQL.ILSIFJob): JSX.Element {
    if (job.type === 'convert') {
        const {
            repository,
            commit,
            root,
        }: {
            repository: string
            commit: string
            root: string
        } = job.arguments

        return (
            <span>
                Convert upload for <strong>{repository}</strong> at{' '}
                <strong>
                    <code>{commit.substring(0, 7)}</code>
                </strong>
                {root !== '' && (
                    <>
                        , <strong>{root}</strong>
                    </>
                )}
            </span>
        )
    }

    const internalJobs: { [K: string]: string } = {
        'clean-old-jobs': 'Purge old job data from LSIF work queue',
        'clean-failed-jobs': 'Clean old failed job uploads from disk',
        'update-tips': 'Refresh current uploads',
    }

    if (internalJobs[job.type]) {
        return (
            <span>
                <strong>Internal job: </strong>
                {internalJobs[job.type]}
            </span>
        )
    }

    return <span>Unknown job type {job.type}</span>
}
