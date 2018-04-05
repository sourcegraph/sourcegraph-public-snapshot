import Loader from '@sourcegraph/icons/lib/Loader'
import upperFirst from 'lodash/upperFirst'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError, ErrorLike, isErrorLike } from '../../util/errors'
import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'
import { RepositoryCompareCommitsPage } from './RepositoryCompareCommitsPage'
import { RepositoryCompareDiffPage } from './RepositoryCompareDiffPage'

function queryRepositoryComparison(args: {
    repo: GQLID
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
            if (!repo.comparison || !repo.comparison.range || errors) {
                throw createAggregateError(errors)
            }
            return repo.comparison.range
        })
    )
}

interface Props extends RepositoryCompareAreaPageProps, RouteComponentProps<{}> {}

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
                        (a, b) =>
                            a.repo.id === b.repo.id &&
                            a.comparisonBaseSpec === b.comparisonBaseSpec &&
                            a.comparisonHeadSpec === b.comparisonHeadSpec
                    ),
                    switchMap(({ repo, comparisonBaseSpec, comparisonHeadSpec }) => {
                        type PartialStateUpdate = Pick<State, 'rangeOrError'>
                        if (!comparisonBaseSpec && !comparisonHeadSpec) {
                            return of({ rangeOrError: null })
                        }
                        return merge(
                            of({ rangeOrError: undefined }),
                            queryRepositoryComparison({
                                repo: repo.id,
                                base: comparisonBaseSpec,
                                head: comparisonHeadSpec,
                            }).pipe(catchError(error => [error]), map(c => ({ rangeOrError: c } as PartialStateUpdate)))
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
                {this.state.rangeOrError === null ? (
                    <p>Select a comparison to begin.</p>
                ) : this.state.rangeOrError === undefined ? (
                    <Loader className="icon-inline" />
                ) : isErrorLike(this.state.rangeOrError) ? (
                    <div className="alert alert-danger mt-2">Error: {upperFirst(this.state.rangeOrError.message)}</div>
                ) : (
                    <>
                        <RepositoryCompareCommitsPage {...this.props} />
                        <div className="mb-3" />
                        <RepositoryCompareDiffPage {...this.props} />
                    </>
                )}
            </div>
        )
    }
}
