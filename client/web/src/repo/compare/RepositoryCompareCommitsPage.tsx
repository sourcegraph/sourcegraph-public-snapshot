import * as React from 'react'

import type { NavigateFunction, Location } from 'react-router-dom'
import { type Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { gql } from '@sourcegraph/http-client'
import { CardHeader, Card } from '@sourcegraph/wildcard'

import { queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import type { GitCommitFields, RepositoryComparisonCommitsResult, Scalars } from '../../graphql-operations'
import { GitCommitNode, type GitCommitNodeProps } from '../commits/GitCommitNode'
import { gitCommitFragment } from '../commits/RepositoryCommitsPage'

import type { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

type RepositoryComparisonRepository = Extract<RepositoryComparisonCommitsResult['node'], { __typename?: 'Repository' }>

function queryRepositoryComparisonCommits(args: {
    repo: Scalars['ID']
    base: string | null
    head: string | null
    first?: number
    path?: string
}): Observable<RepositoryComparisonRepository['comparison']['commits']> {
    return queryGraphQL<RepositoryComparisonCommitsResult>(
        gql`
            query RepositoryComparisonCommits($repo: ID!, $base: String, $head: String, $first: Int, $path: String) {
                node(id: $repo) {
                    ... on Repository {
                        comparison(base: $base, head: $head) {
                            commits(first: $first, path: $path) {
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
            if (!data?.node) {
                throw createAggregateError(errors)
            }
            const repo = data.node as RepositoryComparisonRepository
            if (!repo.comparison?.commits || errors) {
                throw createAggregateError(errors)
            }
            return repo.comparison.commits
        })
    )
}

interface Props extends RepositoryCompareAreaPageProps {
    /** An optional path of a specific file being compared */
    path: string | null

    location: Location
    navigate: NavigateFunction
}

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
                        (a, b) =>
                            a.repo.id === b.repo.id &&
                            a.base.revision === b.base.revision &&
                            a.head.revision === b.head.revision
                    )
                )
                .subscribe(() => this.updates.next())
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-compare-page">
                <Card>
                    <CardHeader>Commits</CardHeader>
                    <FilteredConnection<
                        GitCommitFields,
                        Pick<GitCommitNodeProps, 'className' | 'compact' | 'wrapperElement'>
                    >
                        listClassName="list-group list-group-flush"
                        noun="commit"
                        pluralNoun="commits"
                        compact={true}
                        queryConnection={this.queryCommits}
                        nodeComponent={GitCommitNode}
                        nodeComponentProps={{
                            className: 'list-group-item',
                            compact: true,
                            wrapperElement: 'li',
                        }}
                        defaultFirst={50}
                        hideSearch={true}
                        noSummaryIfAllNodesVisible={true}
                        updates={this.updates}
                    />
                </Card>
            </div>
        )
    }

    private queryCommits = (args: {
        first?: number
    }): Observable<RepositoryComparisonRepository['comparison']['commits']> =>
        queryRepositoryComparisonCommits({
            ...args,
            repo: this.props.repo.id,
            base: this.props.base.revision || null,
            head: this.props.head.revision || null,
            path: this.props.path ?? undefined,
        })
}
