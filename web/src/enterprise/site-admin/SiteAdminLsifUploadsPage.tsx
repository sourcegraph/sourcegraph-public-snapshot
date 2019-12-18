import * as GQL from '../../../../shared/src/graphql/schema'
import React, { FunctionComponent } from 'react'
import { ErrorLike, asError } from '../../../../shared/src/util/errors'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifUploads, fetchLsifUploadStatistics } from './backend'
import { isErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { Link } from '../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription, Observable } from 'rxjs'
import { switchMap, catchError } from 'rxjs/operators'
import { Timestamp } from '../../components/time/Timestamp'
import {
    FilteredConnection,
    FilteredConnectionQueryArgs,
    FilteredConnectionFilter,
} from '../../components/FilteredConnection'
import { ErrorAlert } from '../../components/alerts'

interface LsifUploadNodeProps {
    node: GQL.ILSIFUpload
}

const LsifUploadNode: FunctionComponent<LsifUploadNodeProps> = ({ node }) => (
    <li className="lsif-upload list-group-item">
        <div className="lsif-upload__row lsif-upload__main">
            <div className="lsif-upload__meta">
                <div className="lsif-upload__meta-root">
                    <Link to={`/site-admin/lsif-uploads/${node.id}`}>
                        {node.projectRoot.commit.repository.name}@<code>{node.projectRoot.commit.abbreviatedOID}</code>
                        {node.projectRoot.path === '' ? '' : ` rooted at ${node.projectRoot.path}`}
                    </Link>
                </div>
            </div>

            <small className="text-muted lsif-upload__meta-timestamp">
                <Timestamp noAbout={true} date={node.finishedAt || node.startedAt || node.uploadedAt} />
            </small>
        </div>
    </li>
)

interface Props extends RouteComponentProps<{}> {}

interface State {
    statsOrError: GQL.ILSIFUploadStats | ErrorLike | null
}

/**
 * A page displaying LSIF upload activity.
 */
export class SiteAdminLsifUploadsPage extends React.Component<Props, State> {
    private static FILTERS: FilteredConnectionFilter[] = [
        {
            label: 'Processing',
            id: 'processing',
            tooltip: 'Show processing uploads only',
            args: { state: 'processing' },
        },
        {
            label: 'Errored',
            id: 'errored',
            tooltip: 'Show errored uploads only',
            args: { state: 'errored' },
        },
        {
            label: 'Completed',
            id: 'completed',
            tooltip: 'Show completed uploads only',
            args: { state: 'completed' },
        },
        {
            label: 'Queued',
            id: 'queued',
            tooltip: 'Show queued uploads only',
            args: { state: 'queued' },
        },
    ]

    public state: State = {
        statsOrError: null,
    }

    /** Emits when stats should be refreshed. */
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminLsifUploads')

        this.subscriptions.add(
            this.updates
                // Do not set statsOrError as null here to indicate loading
                // so that the stats don't disappear during navigation, which
                // causes some weird jitter.
                .pipe(switchMap(() => fetchLsifUploadStatistics().pipe(catchError(err => [asError(err)]))))
                .subscribe(statsOrError => this.setState({ statsOrError }))
        )
        this.updates.next()
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-lsif-uploads-page">
                <PageTitle title="LSIF uploads - Admin" />
                <div className="d-flex justify-content-between align-items-center mt-3 mb-1">
                    <h2 className="mb-0">LSIF uploads</h2>
                </div>
                {!this.state.statsOrError ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.statsOrError) ? (
                    <div className="alert alert-danger">
                        <ErrorAlert error={this.state.statsOrError} />
                    </div>
                ) : (
                    <div className="mb-3">
                        {this.state.statsOrError.processingCount} processing, {this.state.statsOrError.erroredCount}{' '}
                        errored, {this.state.statsOrError.completedCount} completed, and{' '}
                        {this.state.statsOrError.queuedCount} queued scheduled
                    </div>
                )}

                <FilteredConnection<{}, LsifUploadNodeProps>
                    noun="upload"
                    pluralNoun="uploads"
                    queryConnection={this.queryUploads}
                    nodeComponent={LsifUploadNode}
                    filters={SiteAdminLsifUploadsPage.FILTERS}
                    history={this.props.history}
                    location={this.props.location}
                    listClassName="list-group list-group-flush mt-3"
                    cursorPaging={true}
                />
            </div>
        )
    }

    private queryUploads = (
        args: FilteredConnectionQueryArgs & { state?: GQL.LSIFUploadState }
    ): Observable<GQL.ILSIFUploadConnection> => {
        // Also refresh stats on each fetch
        this.updates.next()

        // Do the actual query
        return fetchLsifUploads({
            // For typechecker: ensure state is not undefined
            state: args.state || GQL.LSIFUploadState.COMPLETED,
            ...args,
        })
    }
}
