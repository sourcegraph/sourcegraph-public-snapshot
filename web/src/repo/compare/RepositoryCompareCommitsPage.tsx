import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { createAggregateError } from '../../util/errors'
import { GitCommitNode } from '../commits/GitCommitNode'
import { FilteredGitCommitConnection, gitCommitFragment } from '../commits/RepositoryCommitsPage'
import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

export function queryRepositoryComparisonCommits(args: {
    repo: GQLID
    base: string | null
    head: string | null
    first?: number
}): Observable<GQL.IGitCommitConnection> {
    return queryGraphQL(
        gql`
            query RepositoryComparisonCommits($repo: ID!, $base: String, $head: String, $first: Int) {
                node(id: $repo) {
                    ... on Repository {
                        comparison(base: $base, head: $head) {
                            commits(first: $first) {
                                nodes {
                                    ...GitCommitFields
                                }
                                pageInfo {
                                    hasNextPage
                                }
                            }
                        }
                    }
                }
            }
            ${gitCommitFragment}
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node) {
                throw createAggregateError(errors)
            }
            const repo = data.node as GQL.IRepository
            if (!repo.comparison || !repo.comparison.commits || errors) {
                throw createAggregateError(errors)
            }
            return repo.comparison.commits
        })
    )
}

interface Props extends RepositoryCompareAreaPageProps, RouteComponentProps<{}> {}

/** A page with a list of commits in the comparison. */
export class RepositoryCompareCommitsPage extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    distinctUntilChanged(
                        (a, b) => a.repo.id === b.repo.id && a.base.rev === b.base.rev && a.head.rev === b.head.rev
                    )
                )
                .subscribe(() => this.updates.next())
        )
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
                <div className="card">
                    <div className="card-header">Commits</div>
                    <FilteredGitCommitConnection
                        listClassName="list-group list-group-flush"
                        noun="commit"
                        pluralNoun="commits"
                        compact={true}
                        queryConnection={this.queryCommits}
                        nodeComponent={GitCommitNode}
                        nodeComponentProps={{
                            repoName: this.props.repo.uri,
                            className: 'list-group-item',
                            compact: true,
                        }}
                        defaultFirst={50}
                        hideFilter={true}
                        noSummaryIfAllNodesVisible={true}
                        updates={this.updates}
                        history={this.props.history}
                        location={this.props.location}
                    />
                </div>
            </div>
        )
    }

    private queryCommits = (args: { first?: number }): Observable<GQL.IGitCommitConnection> =>
        queryRepositoryComparisonCommits({
            ...args,
            repo: this.props.repo.id,
            base: this.props.base.rev || null,
            head: this.props.head.rev || null,
        })
}
