import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { getDocumentNode, gql } from '@sourcegraph/http-client'
import { IRetentionPolicyOverviewOnLSIFUploadArguments } from '@sourcegraph/shared/src/schema'

import { Connection } from '../../../../components/FilteredConnection'
import {
    GitObjectType,
    LsifUploadRetentionMatchesResult,
    LsifUploadRetentionMatchesVariables,
} from '../../../../graphql-operations'
import { retentionByBranchTipTitle, retentionByUploadTitle } from '../components/UploadRetentionStatusNode'

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
    query LsifUploadRetentionMatches($id: ID!, $matchesOnly: Boolean!, $after: String, $first: Int, $query: String) {
        node(id: $id) {
            __typename
            ... on LSIFUpload {
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

        lsifUploads(dependentOf: $id) {
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
export const queryUploadRetentionMatches = (
    client: ApolloClient<object>,
    id: string,
    { matchesOnly, after, first, query }: IRetentionPolicyOverviewOnLSIFUploadArguments
): Observable<Connection<NormalizedUploadRetentionMatch>> => {
    const vars: LsifUploadRetentionMatchesVariables = {
        id,
        matchesOnly,
        query: query ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<LsifUploadRetentionMatchesResult, LsifUploadRetentionMatchesVariables>({
            query: getDocumentNode(UPLOAD_RETENTIONS_QUERY),
            variables: { ...vars },
        })
    ).pipe(
        map(({ data }) => {
            const { node, ...rest } = data
            if (!node || node.__typename !== 'LSIFUpload') {
                throw new Error('No such LSIFUpload')
            }

            return { node, ...rest }
        }),
        map(({ node, lsifUploads }) => {
            const conn: Connection<NormalizedUploadRetentionMatch> = {
                totalCount:
                    (node.retentionPolicyOverview.totalCount ?? 0) + ((lsifUploads.totalCount ?? 0) > 0 ? 1 : 0),
                nodes: [],
            }

            if ((lsifUploads.totalCount ?? 0) > 0 && retentionByUploadTitle.toLowerCase().includes(query ?? '')) {
                conn.nodes.push({
                    matchType: 'UploadReference',
                    uploadSlice: lsifUploads.nodes,
                    total: lsifUploads.totalCount ?? 0,
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
