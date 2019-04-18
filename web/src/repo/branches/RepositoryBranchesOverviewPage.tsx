import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { upperFirst } from 'lodash'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import * as React from 'react'
import { Link, RouteComponentProps } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { createAggregateError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { memoizeObservable } from '../../../../shared/src/util/memoizeObservable'
import { queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { gitRefFragment, GitRefNode } from '../GitRef'
import { RepositoryBranchesAreaPageProps } from './RepositoryBranchesArea'

interface Data {
    defaultBranch: GQL.IGitRef | null
    activeBranches: GQL.IGitRef[]
    hasMoreActiveBranches: boolean
}

const queryGitBranches = memoizeObservable(
    (args: { repo: GQL.ID; first: number }): Observable<Data> =>
        queryGraphQL(
            gql`
                query RepositoryGitBranchesOverview($repo: ID!, $first: Int!, $withBehindAhead: Boolean!) {
                    node(id: $repo) {
                        ... on Repository {
                            defaultBranch {
                                ...GitRefFields
                            }
                            gitRefs(first: $first, type: GIT_BRANCH, orderBy: AUTHORED_OR_COMMITTED_AT) {
                                nodes {
                                    ...GitRefFields
                                }
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
                ${gitRefFragment}
            `,
            { ...args, withBehindAhead: true }
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.node) {
                    throw createAggregateError(errors)
                }
                const repo = data.node as GQL.IRepository
                if (!repo.gitRefs || !repo.gitRefs.nodes) {
                    throw createAggregateError(errors)
                }
                return {
                    defaultBranch: repo.defaultBranch,
                    activeBranches: repo.gitRefs.nodes.filter(
                        // Filter out default branch from activeBranches.
                        ({ id }) => !repo.defaultBranch || repo.defaultBranch.id !== id
                    ),
                    hasMoreActiveBranches: repo.gitRefs.pageInfo.hasNextPage,
                }
            })
        ),
    args => `${args.repo}:${args.first}`
)

interface Props extends RepositoryBranchesAreaPageProps, RouteComponentProps<{}> {}

interface State {
    /** The page content, undefined while loading, or an error. */
    dataOrError?: Data | ErrorLike
}

/** A page with an overview of the repository's branches. */
export class RepositoryBranchesOverviewPage extends React.PureComponent<Props, State> {
    public state: State = {}

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryBranchesOverview')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    distinctUntilChanged((a, b) => a.repo.id === b.repo.id),
                    switchMap(({ repo }) => {
                        type PartialStateUpdate = Pick<State, 'dataOrError'>
                        return queryGitBranches({ repo: repo.id, first: 10 }).pipe(
                            catchError(error => [error]),
                            map((c): PartialStateUpdate => ({ dataOrError: c })),
                            startWith<PartialStateUpdate>({ dataOrError: undefined })
                        )
                    })
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
        this.componentUpdates.next(this.props)
    }

    public componentWillUpdate(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-branches-page">
                <PageTitle title="Branches" />
                {this.state.dataOrError === undefined ? (
                    <LoadingSpinner className="icon-inline mt-2" />
                ) : isErrorLike(this.state.dataOrError) ? (
                    <div className="alert alert-danger mt-2">Error: {upperFirst(this.state.dataOrError.message)}</div>
                ) : (
                    <div className="repository-branches-page__cards">
                        {this.state.dataOrError.defaultBranch && (
                            <div className="card repository-branches-page__card">
                                <div className="card-header">Default branch</div>
                                <ul className="list-group list-group-flush">
                                    <GitRefNode node={this.state.dataOrError.defaultBranch} />
                                </ul>
                            </div>
                        )}
                        {this.state.dataOrError.activeBranches.length > 0 && (
                            <div className="card repository-branches-page__card">
                                <div className="card-header">Active branches</div>
                                <div className="list-group list-group-flush">
                                    {this.state.dataOrError.activeBranches.map((b, i) => (
                                        <GitRefNode key={i} node={b} />
                                    ))}
                                    {this.state.dataOrError.hasMoreActiveBranches && (
                                        <Link
                                            className="list-group-item list-group-item-action py-2 d-flex"
                                            to={`/${this.props.repo.name}/-/branches/all`}
                                        >
                                            View more branches
                                            <ChevronRightIcon className="icon-inline" />
                                        </Link>
                                    )}
                                </div>
                            </div>
                        )}
                    </div>
                )}
            </div>
        )
    }
}
