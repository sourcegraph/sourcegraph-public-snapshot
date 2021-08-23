import { parseISO } from 'date-fns'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import { requestGraphQL } from '../../../backend/graphql'
import {
    LsifIndexesForRepoResult,
    LsifIndexesForRepoVariables,
    LsifIndexesResult,
    LsifIndexesVariables,
    LsifIndexFields,
    CodeIntelligenceCommitGraphMetadataResult,
    CodeIntelligenceCommitGraphMetadataVariables,
    QueueAutoIndexJobsForRepoResult,
    QueueAutoIndexJobsForRepoVariables,
} from '../../../graphql-operations'
import { lsifIndexFieldsFragment } from '../shared/backend'

/** Return the code intelligence commit graph for the given repository. */
export function fetchCommitGraphMetadata({
    repository,
}: {
    repository: string
}): Observable<{ stale: boolean; updatedAt: Date | null }> {
    const gqlQuery = gql`
        query CodeIntelligenceCommitGraphMetadata($repository: ID!) {
            node(id: $repository) {
                __typename
                ... on Repository {
                    codeIntelligenceCommitGraph {
                        stale
                        updatedAt
                    }
                }
            }
        }
    `

    return requestGraphQL<CodeIntelligenceCommitGraphMetadataResult, CodeIntelligenceCommitGraphMetadataVariables>(
        gqlQuery,
        {
            repository,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('Invalid repository')
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`The given ID is ${node.__typename}, not Repository`)
            }
            if (!node.codeIntelligenceCommitGraph) {
                throw new Error('Missing code intelligence commit graph value')
            }

            return {
                stale: node.codeIntelligenceCommitGraph.stale,
                updatedAt: node.codeIntelligenceCommitGraph.updatedAt
                    ? parseISO(node.codeIntelligenceCommitGraph.updatedAt)
                    : null,
            }
        })
    )
}

export interface IndexConnection {
    nodes: LsifIndexFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

/**
 * Return LSIF indexes. If a repository is given, only indexes for that repository will be returned. Otherwise,
 * indexes across all repositories are returned.
 */
export function fetchLsifIndexes({
    repository,
    query,
    state,
    first,
    after,
}: { repository?: string } & GQL.ILsifIndexesOnRepositoryArguments): Observable<IndexConnection> {
    const vars = {
        repository,
        query: query ?? null,
        state: state ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    if (repository) {
        const gqlQuery = gql`
            query LsifIndexesForRepo(
                $repository: ID!
                $state: LSIFIndexState
                $first: Int
                $after: String
                $query: String
            ) {
                node(id: $repository) {
                    __typename
                    ... on Repository {
                        lsifIndexes(query: $query, state: $state, first: $first, after: $after) {
                            nodes {
                                ...LsifIndexFields
                            }
                            totalCount
                            pageInfo {
                                endCursor
                                hasNextPage
                            }
                        }
                    }
                }
            }

            ${lsifIndexFieldsFragment}
        `

        return requestGraphQL<LsifIndexesForRepoResult, LsifIndexesForRepoVariables>(gqlQuery, {
            ...vars,
            repository,
        }).pipe(
            map(dataOrThrowErrors),
            map(({ node }) => {
                if (!node) {
                    throw new Error('Invalid repository')
                }
                if (node.__typename !== 'Repository') {
                    throw new Error(`The given ID is ${node.__typename}, not Repository`)
                }

                return node.lsifIndexes
            })
        )
    }

    const gqlQuery = gql`
        query LsifIndexes($state: LSIFIndexState, $first: Int, $after: String, $query: String) {
            lsifIndexes(query: $query, state: $state, first: $first, after: $after) {
                nodes {
                    ...LsifIndexFields
                }
                totalCount
                pageInfo {
                    endCursor
                    hasNextPage
                }
            }
        }

        ${lsifIndexFieldsFragment}
    `

    return requestGraphQL<LsifIndexesResult, LsifIndexesVariables>(gqlQuery, vars).pipe(
        map(dataOrThrowErrors),
        map(({ lsifIndexes }) => lsifIndexes)
    )
}

export function enqueueIndexJob(id: string, revision: string): Observable<LsifIndexFields[]> {
    const query = gql`
        mutation QueueAutoIndexJobsForRepo($id: ID!, $rev: String) {
            queueAutoIndexJobsForRepo(repository: $id, rev: $rev) {
                ...LsifIndexFields
            }
        }

        ${lsifIndexFieldsFragment}
    `

    return requestGraphQL<QueueAutoIndexJobsForRepoResult, QueueAutoIndexJobsForRepoVariables>(query, {
        id,
        rev: revision,
    }).pipe(
        map(dataOrThrowErrors),
        map(({ queueAutoIndexJobsForRepo }) => queueAutoIndexJobsForRepo)
    )
}
