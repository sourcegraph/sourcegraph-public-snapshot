import * as GQL from '../../../../shared/src/graphql/schema'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { catchError, map } from 'rxjs/operators'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchLsifDumps } from './backend'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { Link } from '../../../../shared/src/components/Link'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { sortBy } from 'lodash'
import { Timestamp } from '../../components/time/Timestamp'
import { useObservable } from '../../util/useObservable'

const LsifDumpNode: FunctionComponent<{ node: GQL.ILSIFDump }> = ({ node }) => (
    <div className="w-100 lsif-dump list-group-item lsif-dump__main">
        <div className="lsif-dump__meta">
            <div className="lsif-dump__meta-root">
                <Link to={node.projectRoot.url}>
                    <strong>{node.projectRoot.path}</strong>
                </Link>{' '}
                <small className="text-muted lsif-dump__meta-commit">
                    <code>{node.projectRoot.commit.abbreviatedOID}</code>
                </small>
            </div>
        </div>

        <small className="text-muted lsif-dump__meta-timestamp">
            <Timestamp noAbout={true} date={node.uploadedAt} />
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

    const dumpsOrError = useObservable(
        useMemo(
            () =>
                fetchLsifDumps({ repository: repo.id, isLatestForRepo: true, first: 5000 }).pipe(
                    map(({ nodes }: { nodes: GQL.ILSIFDump[] }) => sortBy(nodes, node => node.projectRoot.path)),
                    catchError((error: ErrorLike) => [error])
                ),
            [repo.id]
        )
    )

    const queryDumps = useCallback(
        (args: FilteredConnectionQueryArgs) => fetchLsifDumps({ repository: repo.id, ...args }),
        [repo.id]
    )

    return (
        <div className="repo-settings-code-intelligence-page">
            <PageTitle title="Code intelligence" />
            <h2>Code intelligence</h2>
            <p>
                Enable precise code intelligence by{' '}
                <a href="https://docs.sourcegraph.com/user/code_intelligence/lsif">uploading LSIF data</a>.
            </p>

            <div className="mt-2">
                <h3>Current LSIF uploads</h3>
                <p>
                    These uploads provide code intelligence for the latest commit and are used in cross-repository{' '}
                    <em>Find References</em> requests.
                </p>

                {dumpsOrError === undefined ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(dumpsOrError) ? (
                    <div className="alert alert-danger">
                        Error getting LSIF uploads for repository:
                        <br />
                        <code>{dumpsOrError.message}</code>
                    </div>
                ) : dumpsOrError.length > 0 ? (
                    dumpsOrError.map(dump => <LsifDumpNode key={dump.id} node={dump} />)
                ) : (
                    <p>No uploads are recent enough to be used at the tip of the default branch.</p>
                )}
            </div>

            <div className="mt-2">
                <h3>Historic LSIF uploads</h3>
                <p>These uploads provide code intelligence for older commits.</p>

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
            </div>
        </div>
    )
}
