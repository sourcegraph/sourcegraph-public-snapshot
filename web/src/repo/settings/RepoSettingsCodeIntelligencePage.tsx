import * as GQL from '../../../../shared/src/graphql/schema'
import * as React from 'react'
import { eventLogger } from '../../tracking/eventLogger'
import { PageTitle } from '../../components/PageTitle'
import { RouteComponentProps } from 'react-router'
import { fetchLsifDumps } from './backend'
import { FilteredConnection, FilteredConnectionQueryArgs } from '../../components/FilteredConnection'
import { Timestamp } from '../../components/time/Timestamp'
import { Link } from '../../../../shared/src/components/Link'
import { catchError, map } from 'rxjs/operators'
import { Observable } from 'rxjs'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { sortBy } from 'lodash'
import { ErrorLike } from '@sourcegraph/codeintellify/lib/errors'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { useObservable } from '../../util/useObservable'

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

/**
 * The repository settings code intelligence page.
 */
export const RepoSettingsCodeIntelligencePage: React.FunctionComponent<Props> = ({ repo, ...props }) => {
    React.useEffect(() => eventLogger.logViewEvent('RepoSettingsCodeIntelligence'), [])

    const dumpsOrError = useObservable(
        React.useMemo(
            () =>
                fetchLsifDumps({ repository: repo.id, isLatestForRepo: true, first: 5000 }).pipe(
                    map(({ nodes }: { nodes: GQL.ILSIFDump[] }) => sortBy(nodes, node => node.projectRoot.path)),
                    catchError((error: ErrorLike) => [error])
                ),
            [repo.id]
        )
    )

    const queryDumps = React.useCallback(
        (args: FilteredConnectionQueryArgs): Observable<GQL.ILSIFDumpConnection> =>
            fetchLsifDumps({ repository: repo.id, ...args }),
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

            <div className="lsif-dump-collection">
                <h3>Current LSIF uploads</h3>
                <p>
                    These uploads provide code intelligence for the latest commit and are used in cross-repository{' '}
                    <em>Find Reference</em> requests.
                </p>

                {dumpsOrError === undefined && <LoadingSpinner className="icon-inline" />}
                {dumpsOrError !== undefined && isErrorLike(dumpsOrError) && (
                    <div className="alert alert-danger">
                        Error getting LSIF uploads for repository:
                        <br />
                        <code>{dumpsOrError.message}</code>
                    </div>
                )}
                {dumpsOrError !== undefined && !isErrorLike(dumpsOrError) && dumpsOrError.length > 0 ? (
                    dumpsOrError.map((dump, i) => <LsifDumpNode key={`latest-${dump.id}`} node={dump} />)
                ) : (
                    <p>No uploads are recent enough to be used at the tip of the default branch.</p>
                )}
            </div>

            <div className="lsif-dump-collection">
                <h3>Historic LSIF uploads</h3>
                <p>These uploads provide code intelligence for older commits.</p>

                <FilteredLsifDumpsConnection
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
