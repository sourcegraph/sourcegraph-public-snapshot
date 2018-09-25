import { Hoverifier } from '@sourcegraph/codeintellify'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { merge, Observable, of, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { PageTitle } from '../../components/PageTitle'
import { ExtensionsDocumentsProps } from '../../extensions/environment/ExtensionsEnvironment'
import { ExtensionsControllerProps, ExtensionsProps } from '../../extensions/ExtensionsClientCommonContext'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'
import { RepositoryCompareCommitsPage } from './RepositoryCompareCommitsPage'
import { RepositoryCompareDiffPage } from './RepositoryCompareDiffPage'

function queryRepositoryComparison(args: {
    repo: GQL.ID
    base: string | null
    head: string | null
}): Observable<GQL.IGitRevisionRange> {
    return queryGraphQL(
        gql`
            query RepositoryComparison($repo: ID!, $base: String, $head: String) {
                node(id: $repo) {
                    ... on Repository {
                        comparison(base: $base, head: $head) {
                            range {
                                expr
                                baseRevSpec {
                                    object {
                                        oid
                                    }
                                }
                                headRevSpec {
                                    object {
                                        oid
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node) {
                throw createAggregateError(errors)
            }
            const repo = data.node as GQL.IRepository
            if (
                !repo.comparison ||
                !repo.comparison.range ||
                !repo.comparison.range.baseRevSpec ||
                !repo.comparison.range.baseRevSpec.object ||
                !repo.comparison.range.headRevSpec ||
                !repo.comparison.range.headRevSpec.object ||
                errors
            ) {
                throw createAggregateError(errors)
            }
            eventLogger.log('RepositoryComparisonFetched')
            return repo.comparison.range
        })
    )
}

interface Props
    extends RepositoryCompareAreaPageProps,
        RouteComponentProps<{}>,
        ExtensionsProps,
        ExtensionsControllerProps,
        ExtensionsDocumentsProps {
    /** The base of the comparison. */
    base: { repoPath: string; repoID: GQL.ID; rev?: string | null }

    /** The head of the comparison. */
    head: { repoPath: string; repoID: GQL.ID; rev?: string | null }
    hoverifier: Hoverifier
}

interface State {
    /** The comparison's range, null when no comparison is requested, undefined while loading, or an error. */
    rangeOrError?: null | GQL.IGitRevisionRange | ErrorLike
}

/** A page with an overview of the comparison. */
export class RepositoryCompareOverviewPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryCompareOverview')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged(
                        (a, b) => a.repo.id === b.repo.id && a.base.rev === b.base.rev && a.head.rev === b.head.rev
                    ),
                    switchMap(({ repo, base, head }) => {
                        type PartialStateUpdate = Pick<State, 'rangeOrError'>
                        if (!base.rev && !head.rev) {
                            return of({ rangeOrError: null })
                        }
                        return merge(
                            of({ rangeOrError: undefined }),
                            queryRepositoryComparison({
                                repo: repo.id,
                                base: base.rev || null,
                                head: head.rev || null,
                            }).pipe(
                                catchError(error => [error]),
                                map(c => ({ rangeOrError: c } as PartialStateUpdate))
                            )
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-compare-page">
                <PageTitle title="Compare" />
                {this.state.rangeOrError === null ? (
                    <p>Enter two Git revspecs to compare.</p>
                ) : this.state.rangeOrError === undefined ? (
                    <LoadingSpinner className="icon-inline" />
                ) : isErrorLike(this.state.rangeOrError) ? (
                    <div className="alert alert-danger mt-2">Error: {upperFirst(this.state.rangeOrError.message)}</div>
                ) : (
                    <>
                        <RepositoryCompareCommitsPage {...this.props} />
                        <div className="mb-3" />
                        <RepositoryCompareDiffPage
                            {...this.props}
                            base={{
                                repoPath: this.props.base.repoPath,
                                repoID: this.props.base.repoID,
                                rev: this.props.base.rev || null,
                                commitID: this.state.rangeOrError.baseRevSpec.object!.oid,
                            }}
                            head={{
                                repoPath: this.props.head.repoPath,
                                repoID: this.props.head.repoID,
                                rev: this.props.head.rev || null,
                                commitID: this.state.rangeOrError.headRevSpec.object!.oid,
                            }}
                            extensionsController={this.props.extensionsController}
                            extensionsOnVisibleTextDocumentsChange={this.props.extensionsOnVisibleTextDocumentsChange}
                        />
                    </>
                )}
            </div>
        )
    }
}
