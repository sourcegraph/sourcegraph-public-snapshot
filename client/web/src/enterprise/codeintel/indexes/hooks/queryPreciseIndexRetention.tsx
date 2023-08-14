import type { ApolloClient } from '@apollo/client'
import { from, type Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { getDocumentNode, gql } from '@sourcegraph/http-client'

import type { Connection } from '../../../../components/FilteredConnection'
import type {
    GitObjectType,
    PreciseIndexRetentionResult,
    PreciseIndexRetentionVariables,
} from '../../../../graphql-operations'

export const retentionByUploadTitle = 'Retention by reference'
export const retentionByBranchTipTitle = 'Retention by tip of default branch'

export type NormalizedUploadRetentionMatch = RetentionPolicyMatch | UploadReferenceMatch

export interface RetentionPolicyMatch {
    matchType: 'RetentionPolicy'
    matches: boolean
    protectingCommits: string[]
    configurationPolicy: {
        id: string
        name: string
        type: GitObjectType
        retentionDurationHours: number | null
    } | null
}

export interface UploadReferenceMatch {
    matchType: 'UploadReference'
    uploadSlice: {
        id: string
        inputCommit: string
        inputRoot: string
        projectRoot: {
            repository: { id: string; name: string }
        } | null
    }[]
    total: number
}

const UPLOAD_RETENTIONS_QUERY = gql`
    query PreciseIndexRetention(
        $id: ID!
        $matchesOnly: Boolean!
        $after: String
        $first: Int
        $query: String
        $numReferences: Int
    ) {
        node(id: $id) {
            __typename
            ... on PreciseIndex {
                retentionPolicyOverview(matchesOnly: $matchesOnly, query: $query, after: $after, first: $first) {
                    __typename
                    nodes {
                        __typename
                        configurationPolicy {
                            __typename
                            id
                            name
                            type
                            retentionDurationHours
                        }
                        matches
                        protectingCommits
                    }
                    totalCount
                    pageInfo {
                        endCursor
                        hasNextPage
                    }
                }
            }
        }

        preciseIndexes(dependentOf: $id, first: $numReferences) {
            __typename
            totalCount
            nodes {
                id
                inputCommit
                inputRoot
                projectRoot {
                    repository {
                        name
                        id
                    }
                }
            }
        }
    }
`
export const queryPreciseIndexRetention = (
    client: ApolloClient<object>,
    id: string,
    {
        matchesOnly,
        after,
        first,
        query,
        numReferences,
    }: Partial<PreciseIndexRetentionVariables> & Pick<PreciseIndexRetentionVariables, 'matchesOnly'>
): Observable<Connection<NormalizedUploadRetentionMatch>> => {
    const variables: PreciseIndexRetentionVariables = {
        id,
        matchesOnly,
        query: query ?? null,
        first: first ?? null,
        after: after ?? null,
        numReferences: numReferences ?? null,
    }

    return from(
        client.query<PreciseIndexRetentionResult, PreciseIndexRetentionVariables>({
            query: getDocumentNode(UPLOAD_RETENTIONS_QUERY),
            variables: { ...variables },
        })
    ).pipe(
        map(({ data }) => {
            const { node, ...rest } = data
            if (!node || node.__typename !== 'PreciseIndex') {
                throw new Error('No such precise index')
            }

            return { node, ...rest }
        }),
        map(({ node, preciseIndexes: indexes }) => {
            const conn: Connection<NormalizedUploadRetentionMatch> = {
                totalCount: (node.retentionPolicyOverview.totalCount ?? 0) + ((indexes.totalCount ?? 0) > 0 ? 1 : 0),
                nodes: [],
            }

            if ((indexes.totalCount ?? 0) > 0 && retentionByUploadTitle.toLowerCase().includes(query ?? '')) {
                conn.nodes.push({
                    matchType: 'UploadReference',
                    uploadSlice: indexes.nodes,
                    total: indexes.totalCount ?? 0,
                })
            }

            conn.nodes.push(
                ...node.retentionPolicyOverview.nodes
                    .map(
                        (node): NormalizedUploadRetentionMatch => ({
                            matchType: 'RetentionPolicy',
                            ...node,
                            protectingCommits: node.protectingCommits ?? [],
                        })
                    )
                    .filter(node => {
                        if (
                            node.matchType === 'RetentionPolicy' &&
                            !node.configurationPolicy &&
                            !retentionByBranchTipTitle.toLowerCase().includes(query ?? '')
                        ) {
                            return false
                        }
                        return true
                    })
            )

            conn.pageInfo = node.retentionPolicyOverview.pageInfo
            return conn
        })
    )
}
