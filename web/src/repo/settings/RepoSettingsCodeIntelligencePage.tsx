import * as GQL from '../../../../shared/src/graphql/schema'
import * as React from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { fetchLsifDumps } from './backend'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { Timestamp } from '../../components/time/Timestamp'
import { Link } from '../../../../shared/src/components/Link'
import { switchMap, tap } from 'rxjs/operators'
import { Subject, Subscription, Observable } from 'rxjs'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { sortBy } from 'lodash'

interface LsifDumpNodeProps {
    node: GQL.ILSIFDump
}

class LsifDumpNode extends React.PureComponent<LsifDumpNodeProps, {}> {
    public render(): JSX.Element | null {
        return (
            <div className="lsif-dump list-group-item">
                <div className="lsif-dump__row lsif-dump__main">
                    <div className="lsif-dump__meta">
                        <div className="lsif-dump__meta-root">
                            <Link to={this.props.node.projectRoot.url}>
                                <strong>{this.props.node.projectRoot.path}</strong>
                            </Link>{' '}
                            <small className="text-muted lsif-dump__meta-commit">
                                <code>{this.props.node.projectRoot.commit.abbreviatedOID}</code>
                            </small>
                        </div>
                    </div>

                    <small className="text-muted lsif-dump__meta-timestamp">
                        <Timestamp noAbout={true} date={this.props.node.uploadedAt} />
                    </small>
                </div>
            </div>
        )
    }
}

class FilteredLsifDumpsConnection extends FilteredConnection<{}, LsifDumpNodeProps> {}

interface Props extends RouteComponentProps<any> {
    repo: GQL.IRepository
}

interface State {
    latestDumps?: GQL.ILSIFDump[]
    loading: boolean
    error?: Error
}

/**
 * The repository settings code intelligence page.
 */
export class RepoSettingsCodeIntelligencePage extends React.PureComponent<Props, State> {
    public state: State = { loading: true }

    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepoSettingsCodeIntelligence')

        this.subscriptions.add(
            this.updates
                .pipe(
                    tap(() => this.setState({ loading: true })),
                    switchMap(() => this.queryLatestDumps())
                )
                .subscribe(
                    ({ nodes }: { nodes: GQL.ILSIFDump[] }) =>
                        this.setState({
                            latestDumps: sortBy(nodes, node => node.projectRoot.path),
                            loading: false,
                        }),
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
            <div className="repo-settings-code-intelligence-page">
                <PageTitle title="Code Intelligence" />
                <h2>Code Intelligence</h2>
                <p>
                    Enable precise code intelligence by{' '}
                    <a href="https://docs.sourcegraph.com/user/code_intelligence/lsif">uploading LSIF data</a>.
                </p>

                <div className="lsif-dump-collection">
                    <h3>Current LSIF Uploads</h3>
                    <p>
                        These uploads provide code intelligence for the latest commit and are used in cross-repository{' '}
                        <em>Find Reference</em> requests.
                    </p>

                    {this.state.loading && <LoadingSpinner className="icon-inline" />}
                    {this.state.error && (
                        <div className="alert alert-danger">
                            Error getting repository index status:
                            <br />
                            <code>{this.state.error.message}</code>
                        </div>
                    )}
                    {!this.state.error &&
                    !this.state.loading &&
                    this.state.latestDumps &&
                    this.state.latestDumps.length > 0 ? (
                        this.state.latestDumps.map((dump, i) => <LsifDumpNode key={i} node={dump} />)
                    ) : (
                        <p>No dumps are recent enough to be used at the tip of the default branch.</p>
                    )}
                </div>

                <div className="lsif-dump-collection">
                    <h3>Historic LSIF Uploads</h3>
                    <p>These uploads provide code intelligence for older commits.</p>

                    <FilteredLsifDumpsConnection
                        className="list-group list-group-flush mt-3"
                        noun="upload"
                        pluralNoun="uploads"
                        queryConnection={this.queryDumps}
                        nodeComponent={LsifDumpNode}
                        history={this.props.history}
                        location={this.props.location}
                        listClassName="list-group list-group-flush"
                        cursorPaging={true}
                    />
                </div>
            </div>
        )
    }

    private queryDumps = (args: FilteredConnectionQueryArgs): Observable<GQL.ILSIFDumpConnection> =>
        fetchLsifDumps({ repository: this.props.repo.id, ...args })

    private queryLatestDumps = (): Observable<GQL.ILSIFDumpConnection> =>
        fetchLsifDumps({ repository: this.props.repo.id, isLatestForRepo: true, first: 5000 })
}
