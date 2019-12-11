import * as GQL from '../../../../../shared/src/graphql/schema'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { eventLogger } from '../../../tracking/eventLogger'
import { fetchLsifDumps, fetchLsifUploads } from './backend'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../../components/FilteredConnection'
import { Link } from '../../../../../shared/src/components/Link'
import { PageTitle } from '../../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { Timestamp } from '../../../components/time/Timestamp'
import { Collapsible } from '../../../components/Collapsible'
import { useObservable } from '../../../util/useObservable'
import { Observable } from 'rxjs'

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
            <Timestamp noAbout={true} date={node.processedAt} />
        </small>
    </div>
)

const LsifUploadNode: FunctionComponent<{ node: GQL.ILSIFUpload }> = ({ node }) => (
    <div className="w-100 list-group-item py-2 lsif-data__main">
        <div className="lsif-data__meta">
            <div className="lsif-data__meta-root">
                <code>{node.projectRoot.commit.abbreviatedOID}</code>
                <span className="ml-2">
                    <Link to={node.projectRoot.url}>
                        <strong>{node.projectRoot.path || '/'}</strong>
                    </Link>
                </span>
                <span className="ml-2">
                    -
                    <span className="ml-2">
                        <Link to={`/site-admin/lsif-uploads/${node.id}`}>
                            {node.state === GQL.LSIFUploadState.PROCESSING ? (
                                <span>Processing</span>
                            ) : node.state === GQL.LSIFUploadState.COMPLETED ? (
                                <span className="text-success">Processed</span>
                            ) : node.state === GQL.LSIFUploadState.ERRORED ? (
                                <span className="text-danger">Failed to process</span>
                            ) : (
                                <span>Waiting to process</span>
                            )}
                        </Link>
                    </span>
                </span>
            </div>
        </div>

        <small className="text-muted lsif-data__meta-timestamp">
            <Timestamp noAbout={true} date={node.finishedAt || node.startedAt || node.uploadedAt} />
        </small>
    </div>
)

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

    const fetchUploads = (query: string, state: GQL.LSIFUploadState): Observable<GQL.ILSIFUploadConnection> =>
        fetchLsifUploads({ query, state, first: 5 })

    const activeLsifUploads = useObservable(
        useMemo(() => fetchUploads(repo.name, GQL.LSIFUploadState.PROCESSING), [repo.name])
    )
    const queuedLsifUploads = useObservable(
        useMemo(() => fetchUploads(repo.name, GQL.LSIFUploadState.QUEUED), [repo.name])
    )
    const failedLsifUploads = useObservable(
        useMemo(() => fetchUploads(repo.name, GQL.LSIFUploadState.ERRORED), [repo.name])
    )

    const activeCount = activeLsifUploads?.nodes.length || 0
    const queuedCount = queuedLsifUploads?.nodes.length || 0
    const failedCount = failedLsifUploads?.nodes.length || 0

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
                    useURLQuery={false}
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
                    title="Activity for this repository"
                    defaultExpanded={false}
                    className="repo-settings-code-intelligence-page-collapsible"
                    titleClassName="h5 mb-0"
                >
                    <div>
                        <h3>Pending and active LSIF uploads</h3>

                        {activeLsifUploads === undefined ||
                        queuedLsifUploads === undefined ||
                        activeCount + queuedCount === 0 ? (
                            <p>
                                <small>No uploads are queued or currently being processed.</small>
                            </p>
                        ) : (
                            <>
                                <p>These uploads have been accepted but have not yet been processed.</p>

                                <div className="list-group list-group-flush mt-3">
                                    {activeLsifUploads?.nodes.map(upload => (
                                        <LsifUploadNode key={upload.id} node={upload} />
                                    ))}
                                    {queuedLsifUploads?.nodes.map(upload => (
                                        <LsifUploadNode key={upload.id} node={upload} />
                                    ))}
                                </div>

                                {queuedLsifUploads?.pageInfo.hasNextPage && (
                                    <div className="mt-2">
                                        Showing five queued uploads.{' '}
                                        <Link
                                            to={`http://localhost:3080/site-admin/lsif-uploads?filter=queued&query=${encodeURIComponent(
                                                repo.name
                                            )}`}
                                        >
                                            See all queued uploads for this repository.
                                        </Link>
                                    </div>
                                )}
                            </>
                        )}
                    </div>

                    <div className="mt-4">
                        <h3>Recent failed LSIF uploads</h3>

                        {failedLsifUploads === undefined || failedCount === 0 ? (
                            <p>
                                <small>No recent uploads have failed processing.</small>
                            </p>
                        ) : (
                            <>
                                <p>These uploads have recently failed processing.</p>

                                <div className="list-group list-group-flush mt-3">
                                    {failedLsifUploads.nodes.map(upload => (
                                        <LsifUploadNode key={upload.id} node={upload} />
                                    ))}
                                </div>

                                {failedLsifUploads?.pageInfo.hasNextPage && (
                                    <div className="mt-2">
                                        Showing five recent failures.{' '}
                                        <Link
                                            to={`http://localhost:3080/site-admin/lsif-uploads?filter=errored&query=${encodeURIComponent(
                                                repo.name
                                            )}`}
                                        >
                                            See all failed uploads for this repository.
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
