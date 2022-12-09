import * as React from 'react'

import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { HoverMerged } from '@sourcegraph/client-api'
import { Hoverifier } from '@sourcegraph/codeintellify'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { FileSpec, RepoSpec, ResolvedRevisionSpec, RevisionSpec } from '@sourcegraph/shared/src/util/url'

import { fileDiffFields, diffStatFields } from '../../backend/diff'
import { requestGraphQL } from '../../backend/graphql'
import { FileDiffConnection } from '../../components/diff/FileDiffConnection'
import { FileDiffNode } from '../../components/diff/FileDiffNode'
import { ConnectionQueryArguments } from '../../components/FilteredConnection'
import { RepositoryComparisonDiffResult, RepositoryComparisonDiffVariables } from '../../graphql-operations'

import { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

export type RepositoryComparisonDiff = Extract<RepositoryComparisonDiffResult['node'], { __typename?: 'Repository' }>

export function queryRepositoryComparisonFileDiffs(args: {
    repo: Scalars['ID']
    base: string | null
    head: string | null
    first: number | null
    after: string | null
    paths: string[] | null
}): Observable<RepositoryComparisonDiff['comparison']['fileDiffs']> {
    return requestGraphQL<RepositoryComparisonDiffResult, RepositoryComparisonDiffVariables>(
        gql`
            query RepositoryComparisonDiff(
                $repo: ID!
                $base: String
                $head: String
                $first: Int
                $after: String
                $paths: [String!]
            ) {
                node(id: $repo) {
                    __typename
                    ... on Repository {
                        comparison(base: $base, head: $head) {
                            fileDiffs(first: $first, after: $after, paths: $paths) {
                                nodes {
                                    ...FileDiffFields
                                }
                                totalCount
                                pageInfo {
                                    endCursor
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

            ${fileDiffFields}

            ${diffStatFields}
        `,
        args
    ).pipe(
        map(result => {
            const data = dataOrThrowErrors(result)

            const repo = data.node
            if (repo === null) {
                throw new Error('Repository not found')
            }
            if (repo.__typename !== 'Repository') {
                throw new Error('Not a repository')
            }
            return repo.comparison.fileDiffs
        })
    )
}

interface RepositoryCompareDiffPageProps
    extends RepositoryCompareAreaPageProps,
        RouteComponentProps<{}>,
        PlatformContextProps,
        ExtensionsControllerProps,
        ThemeProps {
    /** The base of the comparison. */
    base: { repoName: string; repoID: Scalars['ID']; revision: string | null; commitID: string }

    /** The head of the comparison. */
    head: { repoName: string; repoID: Scalars['ID']; revision: string | null; commitID: string }

    /** An optional path of a specific file to compare */
    path: string | null

    hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
}

/** A page with the file diffs in the comparison. */
export class RepositoryCompareDiffPage extends React.PureComponent<RepositoryCompareDiffPageProps> {
    public render(): JSX.Element | null {
        const { extensionsController } = this.props
        return (
            <div className="repository-compare-page">
                <FileDiffConnection
                    listClassName="list-group list-group-flush test-file-diff-connection"
                    noun="changed file"
                    pluralNoun="changed files"
                    queryConnection={this.queryDiffs}
                    nodeComponent={FileDiffNode}
                    nodeComponentProps={
                        extensionsController !== null
                            ? {
                                  ...this.props,
                                  extensionInfo: {
                                      base: { ...this.props.base, revision: this.props.base.revision || 'HEAD' },
                                      head: { ...this.props.head, revision: this.props.head.revision || 'HEAD' },
                                      hoverifier: this.props.hoverifier,
                                      extensionsController,
                                  },
                                  lineNumbers: true,
                              }
                            : undefined
                    }
                    defaultFirst={15}
                    hideSearch={true}
                    noSummaryIfAllNodesVisible={true}
                    history={this.props.history}
                    location={this.props.location}
                    cursorPaging={true}
                />
            </div>
        )
    }

    private queryDiffs = (
        args: ConnectionQueryArguments
    ): Observable<RepositoryComparisonDiff['comparison']['fileDiffs']> =>
        queryRepositoryComparisonFileDiffs({
            first: args.first ?? null,
            after: args.after ?? null,
            repo: this.props.repo.id,
            base: this.props.base.commitID,
            head: this.props.head.commitID,
            // All of our user journeys are designed for just a single file path, so the component APIs are set up to
            // enforce that, even though the GraphQL query is able to support any number of paths
            paths: this.props.path ? [this.props.path] : [],
        })
}
