import React, { useCallback } from 'react'

import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import type { Scalars } from '@sourcegraph/shared/src/graphql-operations'

import { fileDiffFields, diffStatFields } from '../../backend/diff'
import { requestGraphQL } from '../../backend/graphql'
import { FileDiffNode, type FileDiffNodeProps } from '../../components/diff/FileDiffNode'
import { type ConnectionQueryArguments, FilteredConnection } from '../../components/FilteredConnection'
import type {
    RepositoryComparisonDiffResult,
    RepositoryComparisonDiffVariables,
    FileDiffFields,
} from '../../graphql-operations'

import type { RepositoryCompareAreaPageProps } from './RepositoryCompareArea'

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

interface RepositoryCompareDiffPageProps extends RepositoryCompareAreaPageProps {
    /** The base of the comparison. */
    base: { repoName: string; repoID: Scalars['ID']; revision: string | null; commitID: string }

    /** The head of the comparison. */
    head: { repoName: string; repoID: Scalars['ID']; revision: string | null; commitID: string }

    /** An optional path of a specific file to compare */
    path: string | null
}

/** A page with the file diffs in the comparison. */
export const RepositoryCompareDiffPage: React.FunctionComponent<RepositoryCompareDiffPageProps> = props => {
    const queryDiffs = useCallback(
        (args: ConnectionQueryArguments): Observable<RepositoryComparisonDiff['comparison']['fileDiffs']> =>
            queryRepositoryComparisonFileDiffs({
                first: args.first ?? null,
                after: args.after ?? null,
                repo: props.repo.id,
                base: props.base.commitID,
                head: props.head.commitID,
                // All of our user journeys are designed for just a single file path, so the component APIs are set up to
                // enforce that, even though the GraphQL query is able to support any number of paths
                paths: props.path ? [props.path] : [],
            }),
        [props.base.commitID, props.head.commitID, props.path, props.repo.id]
    )

    return (
        <div className="repository-compare-page">
            <FilteredConnection<FileDiffFields, Omit<FileDiffNodeProps, 'node'>>
                listClassName="list-group list-group-flush test-file-diff-connection"
                noun="changed file"
                pluralNoun="changed files"
                queryConnection={queryDiffs}
                nodeComponent={FileDiffNode}
                nodeComponentProps={{ ...props, lineNumbers: true }}
                defaultFirst={15}
                hideSearch={true}
                noSummaryIfAllNodesVisible={true}
                withCenteredSummary={true}
                cursorPaging={true}
            />
        </div>
    )
}
