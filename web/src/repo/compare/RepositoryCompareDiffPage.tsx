import { Hoverifier } from '@sourcegraph/codeintellify'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { ActionItemProps } from '../../../../shared/src/actions/ActionItem'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { gql } from '../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { createAggregateError } from '../../../../shared/src/util/errors'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '../../../../shared/src/util/url'
import { queryGraphQL } from '../../backend/graphql'
import { FileDiffConnection } from './FileDiffConnection'
import { FileDiffNode } from './FileDiffNode'
import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

export function queryRepositoryComparisonFileDiffs(args: {
    repo: GQL.ID
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
                mostRelevantFile {
                    url
                }
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
                internalID
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

interface RepositoryCompareDiffPageProps
    extends RepositoryCompareAreaPageProps,
        RouteComponentProps<{}>,
        PlatformContextProps,
        ExtensionsControllerProps {
    /** The base of the comparison. */
    base: { repoName: string; repoID: GQL.ID; rev: string | null; commitID: string }

    /** The head of the comparison. */
    head: { repoName: string; repoID: GQL.ID; rev: string | null; commitID: string }
    hoverifier: Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemProps>
}

/** A page with the file diffs in the comparison. */
export class RepositoryCompareDiffPage extends React.PureComponent<RepositoryCompareDiffPageProps> {
    public render(): JSX.Element | null {
        return (
            <div className="repository-compare-page">
                <FileDiffConnection
                    listClassName="list-group list-group-flush"
                    noun="changed file"
                    pluralNoun="changed files"
                    queryConnection={this.queryDiffs}
                    nodeComponent={FileDiffNode}
                    nodeComponentProps={{
                        base: { ...this.props.base, rev: this.props.base.rev || 'HEAD' },
                        head: { ...this.props.head, rev: this.props.head.rev || 'HEAD' },
                        lineNumbers: true,
                        platformContext: this.props.platformContext,
                        location: this.props.location,
                        history: this.props.history,
                        hoverifier: this.props.hoverifier,
                        extensionsController: this.props.extensionsController,
                    }}
                    defaultFirst={25}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                    extensionsController={this.props.extensionsController}
                />
            </div>
        )
    }

    private queryDiffs = (args: { first?: number }): Observable<GQL.IFileDiffConnection> =>
        queryRepositoryComparisonFileDiffs({
            ...args,
            repo: this.props.repo.id,
            base: this.props.base.commitID,
            head: this.props.head.commitID,
        })
}
