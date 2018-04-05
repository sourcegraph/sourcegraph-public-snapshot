import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs/Observable'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { gql, queryGraphQL } from '../../backend/graphql'
import { FilteredConnection } from '../../components/FilteredConnection'
import { eventLogger } from '../../tracking/eventLogger'
import { createAggregateError } from '../../util/errors'
import { FileDiffNode, FileDiffNodeProps } from './FileDiffNode'
import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

function queryRepositoryComparisonFileDiffs(args: {
    repo: GQLID
    base: string | null
    head: string | null
    first?: number
}): Observable<GQL.IFileDiffConnection> {
    return queryGraphQL(
        gql`
            query RepositoryComparisonDiff($repo: ID!, $base: String, $head: String, $first: Int) {
                node(id: $repo) {
                    ... on Repository {
                        comparison(base: $base, head: $head) {
                            fileDiffs(first: $first) {
                                nodes {
                                    ...FileDiffFields
                                }
                                totalCount
                                pageInfo {
                                    hasNextPage
                                }
                                diffStat {
                                    ...DiffStatFields
                                }
                            }
                        }
                    }
                }
            }

            fragment FileDiffFields on FileDiff {
                oldPath
                newPath
                hunks {
                    oldRange {
                        ...FileDiffHunkRangeFields
                    }
                    oldNoNewlineAt
                    newRange {
                        ...FileDiffHunkRangeFields
                    }
                    section
                    body
                }
                stat {
                    ...DiffStatFields
                }
            }

            fragment FileDiffHunkRangeFields on FileDiffHunkRange {
                startLine
                lines
            }

            fragment DiffStatFields on DiffStat {
                added
                changed
                deleted
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.node) {
                throw createAggregateError(errors)
            }
            const repo = data.node as GQL.IRepository
            if (!repo.comparison || !repo.comparison.fileDiffs || errors) {
                throw createAggregateError(errors)
            }
            return repo.comparison.fileDiffs
        })
    )
}

interface Props extends RepositoryCompareAreaPageProps, RouteComponentProps<{}> {}

export class FilteredFileDiffConnection extends FilteredConnection<
    GQL.IFileDiff,
    Pick<FileDiffNodeProps, 'repoName' | 'base' | 'head' | 'className'>
> {}

/** A page with the file diffs in the comparison. */
export class RepositoryCompareDiffPage extends React.PureComponent<Props> {
    private componentUpdates = new Subject<Props>()
    private updates = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        eventLogger.logViewEvent('RepositoryCompareDiff')

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(this.props),
                    distinctUntilChanged(
                        (a, b) =>
                            a.repo.id === b.repo.id &&
                            a.comparisonBaseSpec === b.comparisonBaseSpec &&
                            a.comparisonHeadSpec === b.comparisonHeadSpec
                    )
                )
                .subscribe(() => this.updates.next())
        )
    }

    public componentWillUpdate(nextProps: Props): void {
        this.componentUpdates.next(nextProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="repository-compare-page">
                <FilteredFileDiffConnection
                    listClassName="list-group list-group-flush"
                    noun="changed file"
                    pluralNoun="changed files"
                    queryConnection={this.queryDiffs}
                    nodeComponent={FileDiffNode}
                    nodeComponentProps={{
                        repoName: this.props.repo.uri,
                        base: this.props.comparisonBaseSpec || 'HEAD',
                        head: this.props.comparisonHeadSpec || 'HEAD',
                    }}
                    defaultFirst={5}
                    hideFilter={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                />
            </div>
        )
    }

    private queryDiffs = (args: { first?: number }): Observable<GQL.IFileDiffConnection> =>
        queryRepositoryComparisonFileDiffs({
            ...args,
            repo: this.props.repo.id,
            base: this.props.comparisonBaseSpec,
            head: this.props.comparisonHeadSpec,
        })
}
