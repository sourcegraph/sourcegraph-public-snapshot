import * as React from 'react'

import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import { RouteComponentProps } from 'react-router-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, map, startWith, switchMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError, createAggregateError, ErrorLike, isErrorLike, memoizeObservable } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import * as GQL from '@sourcegraph/shared/src/schema'
import { Link, LoadingSpinner, CardHeader, Card, Icon } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { PageTitle } from '../../components/PageTitle'
import { eventLogger } from '../../tracking/eventLogger'
import { gitReferenceFragments, GitReferenceNode } from '../GitReference'

import { RepositoryBranchesAreaPageProps } from './RepositoryBranchesArea'

import styles from './RepositoryBranchesOverviewPage.module.scss'

interface Data {
    defaultBranch: GQL.IGitRef | null
    activeBranches: GQL.IGitRef[]
    hasMoreActiveBranches: boolean
}

export const queryGitBranches = memoizeObservable(
    (args: { repo: Scalars['ID']; first: number }): Observable<Data> =>
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
                ${gitReferenceFragments}
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
                            catchError((error): [ErrorLike] => [asError(error)]),
                            map((dataOrError): PartialStateUpdate => ({ dataOrError })),
                            startWith<PartialStateUpdate>({ dataOrError: undefined })
                        )
                    })
                )
                .subscribe(
                    stateUpdate => this.setState(stateUpdate),
                    error => console.error(error)
                )
        )
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div>
                <PageTitle title="Branches" />
                {this.state.dataOrError === undefined ? (
                    <LoadingSpinner className="mt-2" />
                ) : isErrorLike(this.state.dataOrError) ? (
                    <ErrorAlert className="mt-2" error={this.state.dataOrError} />
                ) : (
                    <div>
                        {this.state.dataOrError.defaultBranch && (
                            <Card className={styles.card}>
                                <CardHeader>Default branch</CardHeader>
                                <ul className="list-group list-group-flush">
                                    <GitReferenceNode node={this.state.dataOrError.defaultBranch} />
                                </ul>
                            </Card>
                        )}
                        {this.state.dataOrError.activeBranches.length > 0 && (
                            <Card className={styles.card}>
                                <CardHeader>Active branches</CardHeader>
                                <ul className="list-group list-group-flush" data-testid="active-branches-list">
                                    {this.state.dataOrError.activeBranches.map((gitReference, index) => (
                                        <GitReferenceNode key={index} node={gitReference} />
                                    ))}
                                    {this.state.dataOrError.hasMoreActiveBranches && (
                                        <Link
                                            className="list-group-item list-group-item-action py-2 d-flex"
                                            to={`/${this.props.repo.name}/-/branches/all`}
                                        >
                                            View more branches
                                            <Icon role="img" as={ChevronRightIcon} aria-hidden={true} />
                                        </Link>
                                    )}
                                </ul>
                            </Card>
                        )}
                    </div>
                )}
            </div>
        )
    }
}
