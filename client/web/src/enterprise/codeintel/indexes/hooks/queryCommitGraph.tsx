import { ApolloClient } from '@apollo/client'
import { parseISO } from 'date-fns'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { gql, getDocumentNode } from '@sourcegraph/http-client'

import {
    CodeIntelligenceCommitGraphMetadataResult,
    CodeIntelligenceCommitGraphMetadataVariables,
} from '../../../../graphql-operations'

const GRAPH_METADATA = gql`
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

export const queryCommitGraph = (
    repository: string,
    client: ApolloClient<object>
): Observable<{ stale: boolean; updatedAt: Date | null }> =>
    from(
        client.query<CodeIntelligenceCommitGraphMetadataResult, CodeIntelligenceCommitGraphMetadataVariables>({
            query: getDocumentNode(GRAPH_METADATA),
            variables: { repository },
        })
    ).pipe(
        map(({ data }) => data),
        map(({ node }) => {
            if (!node) {
                throw new Error('Invalid repository')
            }
            if (node.__typename !== 'Repository') {
                throw new Error(`The given ID is ${node.__typename}, not Repository`)
            }
            if (!node.codeIntelligenceCommitGraph) {
                throw new Error('Missing code navigation commit graph value')
            }

            return {
                stale: node.codeIntelligenceCommitGraph.stale,
                updatedAt: node.codeIntelligenceCommitGraph.updatedAt
                    ? parseISO(node.codeIntelligenceCommitGraph.updatedAt)
                    : null,
            }
        })
    )
