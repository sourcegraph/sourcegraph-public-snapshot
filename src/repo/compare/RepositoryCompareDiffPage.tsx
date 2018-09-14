import { Hoverifier } from '@sourcegraph/codeintellify'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../../backend/graphql'
import * as GQL from '../../backend/graphqlschema'
import { ExtensionsDocumentsProps } from '../../extensions/environment/ExtensionsEnvironment'
import { ExtensionsControllerProps, ExtensionsProps } from '../../extensions/ExtensionsClientCommonContext'
import { createAggregateError } from '../../util/errors'
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
        ExtensionsProps,
        ExtensionsControllerProps,
        ExtensionsDocumentsProps {
    /** The base of the comparison. */
    base: { repoPath: string; repoID: GQL.ID; rev: string | null; commitID: string }

    /** The head of the comparison. */
    head: { repoPath: string; repoID: GQL.ID; rev: string | null; commitID: string }
    hoverifier: Hoverifier
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
                        extensions: this.props.extensions,
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
                    extensionsOnVisibleTextDocumentsChange={this.props.extensionsOnVisibleTextDocumentsChange}
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
