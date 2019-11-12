import * as GQL from '../../../../shared/src/graphql/schema'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifDumps, fetchLsifJobs } from './backend'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { Link } from '../../../../shared/src/components/Link'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../components/time/Timestamp'
import { Collapsible } from '../../components/Collapsible'
import { useObservable } from '../../util/useObservable'
import { Observable } from 'rxjs'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { ErrorAlert } from '../../components/alerts'

const LsifDumpNode: FunctionComponent<{ node: GQL.ILSIFDump }> = ({ node }) => (
    <div className="w-100 list-group-item py-2 lsif-data__main">
        <div className="lsif-data__meta">
            <div className="lsif-data__meta-root">
                <code>{node.projectRoot.commit.abbreviatedOID}</code>
                <span className="ml-2">
                    <Link to={node.projectRoot.url}>
                        <strong>{node.projectRoot.path || '/'}</strong>
                    </Link>
                </span>
            </div>
        </div>

        <small className="text-muted lsif-data__meta-timestamp">
            <Timestamp noAbout={true} date={node.uploadedAt} />
        </small>
    </div>
)

const LsifJobNode: FunctionComponent<{ node: GQL.ILSIFJob }> = ({ node }) => {
    const { commit, root } = node.args as { commit: string; root: string }

    return (
        <div className="w-100 list-group-item py-2 lsif-data__main">
            <div className="lsif-data__meta">
                <div className="lsif-data__meta-root">
                    <Link to={`/site-admin/lsif-jobs/${node.id}`}>
                        {node.state === GQL.LSIFJobState.PROCESSING ? (
                            <span>Processing</span>
                        ) : node.state === GQL.LSIFJobState.COMPLETED ? (
                            <span className="text-success">Processed</span>
                        ) : node.state === GQL.LSIFJobState.ERRORED ? (
                            <span className="text-danger">Failed to process</span>
                        ) : (
                            <span>Waiting to process</span>
                        )}
                    </Link>{' '}
                    <code>{commit.substring(0, 7)}</code>
                    <span className="ml-2">
                        <strong>{root || '/'}</strong>
                    </span>
                </div>
            </div>

            <small className="text-muted lsif-data__meta-timestamp">
                <Timestamp noAbout={true} date={node.finishedOn || node.processedOn || node.timestamp} />
            </small>
        </div>
    )
}

interface Props extends RouteComponentProps<any> {
    repo: GQL.IRepository
}

/**
 * The repository settings code intelligence page.
 */
export const RepoSettingsCodeIntelligencePage: FunctionComponent<Props> = ({ repo, ...props }) => {
    useEffect(() => eventLogger.logViewEvent('RepoSettingsCodeIntelligence'), [])

    const queryDumps = useCallback(
        (args: FilteredConnectionQueryArgs) => fetchLsifDumps({ repository: repo.id, ...args }),
        [repo.id]
    )

    const queryLatestDumps = useCallback(
        (args: FilteredConnectionQueryArgs) => fetchLsifDumps({ repository: repo.id, isLatestForRepo: true, ...args }),
        [repo.id]
    )

    const fromObservable = useCallback(
        (state: GQL.LSIFJobState): Observable<GQL.ILSIFJobConnection | ErrorLike> =>
            fetchLsifJobs({
                query: `convert ${repo.name}`,
                state,
                first: 5,
            }),
        [repo.name]
    )

    const activeLsifJobs = useObservable(useMemo(() => fromObservable(GQL.LSIFJobState.PROCESSING), [fromObservable]))
    const queuedLsifJobs = useObservable(useMemo(() => fromObservable(GQL.LSIFJobState.QUEUED), [fromObservable]))
    const failedLsifJobs = useObservable(useMemo(() => fromObservable(GQL.LSIFJobState.ERRORED), [fromObservable]))

    const activeCount = (activeLsifJobs && !isErrorLike(activeLsifJobs) && activeLsifJobs.nodes.length) || 0
    const queuedCount = (queuedLsifJobs && !isErrorLike(queuedLsifJobs) && queuedLsifJobs.nodes.length) || 0
    const failedCount = (failedLsifJobs && !isErrorLike(failedLsifJobs) && failedLsifJobs.nodes.length) || 0

    const activityTitle = (
        <>
            <span className="h5">Activity for this repository</span>

            <div className="repo-settings-code-intelligence-page-collapsible-badges ml-1">
                {activeCount + queuedCount > 0 && (
                    <span className="badge badge-primary badge-pill ml-1">
                        {activeCount + queuedCount}
                        {queuedLsifJobs && !isErrorLike(queuedLsifJobs) && queuedLsifJobs.pageInfo.hasNextPage && '+'}
                    </span>
                )}
                {failedCount > 0 && (
                    <span className="badge badge-danger badge-pill ml-1">
                        {failedCount}
                        {failedLsifJobs && !isErrorLike(failedLsifJobs) && failedLsifJobs.pageInfo.hasNextPage && '+'}
                    </span>
                )}
            </div>
        </>
    )

    return (
        <div className="repo-settings-code-intelligence-page">
            <PageTitle title="Code intelligence" />
            <h2>Code intelligence</h2>
            <p>
                Enable precise code intelligence by{' '}
                <a href="https://docs.sourcegraph.com/user/code_intelligence/lsif">uploading LSIF data</a>.
            </p>

            <div className="mt-4">
                <h3>Current LSIF uploads</h3>
                <p>
                    These uploads provide code intelligence for the latest commit and are used in cross-repository{' '}
                    <em>Find References</em> requests.
                </p>

                <FilteredConnection<{}, { node: GQL.ILSIFDump }>
                    className="list-group list-group-flush mt-3"
                    noun="upload"
                    pluralNoun="uploads"
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    queryConnection={queryLatestDumps}
                    nodeComponent={LsifDumpNode}
                    history={props.history}
                    location={props.location}
                    listClassName="list-group list-group-flush"
                    cursorPaging={true}
                    emptyElement={
                        <small>No dumps are recent enough to be used at the tip of the default branch.</small>
                    }
                />
            </div>

            <div className="mt-4">
                <Collapsible
                    title={activityTitle}
                    defaultExpanded={false}
                    className="repo-settings-code-intelligence-page-collapsible"
                    titleClassName="mb-0"
                >
                    <div>
                        <h3>Pending and active LSIF uploads</h3>

                        {activeLsifJobs === undefined || queuedLsifJobs === undefined ? (
                            <></>
                        ) : isErrorLike(activeLsifJobs) ? (
                            <ErrorAlert prefix="Error loading active LSIF jobs" error={activeLsifJobs} />
                        ) : isErrorLike(queuedLsifJobs) ? (
                            <ErrorAlert prefix="Error loading queued LSIF jobs" error={queuedLsifJobs} />
                        ) : activeCount + queuedCount === 0 ? (
                            <p>
                                <small>No uploads are queued or currently being processed.</small>
                            </p>
                        ) : (
                            <>
                                <p>These uploads have been accepted but have not yet been processed.</p>

                                <div className="list-group list-group-flush mt-3">
                                    {activeLsifJobs &&
                                        !isErrorLike(activeLsifJobs) &&
                                        activeLsifJobs.nodes.map(job => <LsifJobNode key={job.id} node={job} />)}
                                    {queuedLsifJobs &&
                                        !isErrorLike(queuedLsifJobs) &&
                                        queuedLsifJobs.nodes.map(job => <LsifJobNode key={job.id} node={job} />)}
                                </div>

                                {queuedLsifJobs && !isErrorLike(queuedLsifJobs) && queuedLsifJobs.pageInfo.hasNextPage && (
                                    <div className="mt-2">
                                        Showing five queued uploads.{' '}
                                        <Link
                                            to={`http://localhost:3080/site-admin/lsif-jobs?filter=queued&query=convert+${encodeURIComponent(
                                                repo.name
                                            )}`}
                                        >
                                            See all queued jobs for this repository.
                                        </Link>
                                    </div>
                                )}
                            </>
                        )}
                    </div>

                    <div className="mt-4">
                        <h3>Recent failed LSIF uploads</h3>

                        {failedLsifJobs === undefined ? (
                            <></>
                        ) : isErrorLike(failedLsifJobs) ? (
                            <ErrorAlert prefix="Error loading failed LSIF jobs" error={failedLsifJobs} />
                        ) : failedCount === 0 ? (
                            <p>
                                <small>No recent uploads have failed processing.</small>
                            </p>
                        ) : (
                            <>
                                <p>These uploads have recently failed processing.</p>

                                <div className="list-group list-group-flush mt-3">
                                    {failedLsifJobs.nodes.map(job => (
                                        <LsifJobNode key={job.id} node={job} />
                                    ))}
                                </div>

                                {failedLsifJobs && failedLsifJobs.pageInfo.hasNextPage && (
                                    <div className="mt-2">
                                        Showing five recent failures.{' '}
                                        <Link
                                            to={`http://localhost:3080/site-admin/lsif-jobs?filter=errored&query=convert+${encodeURIComponent(
                                                repo.name
                                            )}`}
                                        >
                                            See all failed jobs for this repository.
                                        </Link>
                                    </div>
                                )}
                            </>
                        )}
                    </div>
                </Collapsible>
            </div>

            <div className="mt-4">
                <Collapsible
                    title="Historic LSIF uploads"
                    defaultExpanded={false}
                    className="repo-settings-code-intelligence-page-collapsible"
                    titleClassName="h5 mb-0"
                >
                    <p>These uploads provide code intelligence for branches and older commits.</p>

                    <FilteredConnection<{}, { node: GQL.ILSIFDump }>
                        className="list-group list-group-flush mt-3"
                        noun="upload"
                        pluralNoun="uploads"
                        queryConnection={queryDumps}
                        nodeComponent={LsifDumpNode}
                        history={props.history}
                        location={props.location}
                        listClassName="list-group list-group-flush"
                        cursorPaging={true}
                    />
                </Collapsible>
            </div>
        </div>
    )
}
