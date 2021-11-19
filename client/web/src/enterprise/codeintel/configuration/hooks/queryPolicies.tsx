import { ApolloClient } from '@apollo/client'
import { from, Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { getDocumentNode, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

import {
    CodeIntelligenceConfigurationPoliciesResult,
    CodeIntelligenceConfigurationPoliciesVariables,
    CodeIntelligenceConfigurationPolicyFields,
    GitObjectType,
} from '../../../../graphql-operations'

const defaultCodeIntelligenceConfigurationPolicyFieldsFragment = gql`
    fragment CodeIntelligenceConfigurationPolicyFields on CodeIntelligenceConfigurationPolicy {
        __typename
        id
        name
        repository {
            id
            name
        }
        repositoryPatterns
        type
        pattern
        protected
        retentionEnabled
        retentionDurationHours
        retainIntermediateCommits
        indexingEnabled
        indexCommitMaxAgeHours
        indexIntermediateCommits
    }
`

export const nullPolicy = {
    __typename: 'CodeIntelligenceConfigurationPolicy' as const,
    id: '',
    name: '',
    repositoryPatterns: null,
    type: GitObjectType.GIT_UNKNOWN,
    pattern: '',
    protected: false,
    retentionEnabled: false,
    retentionDurationHours: null,
    retainIntermediateCommits: false,
    indexingEnabled: false,
    indexCommitMaxAgeHours: null,
    indexIntermediateCommits: false,
    repository: null,
}

interface PolicyConnection {
    nodes: CodeIntelligenceConfigurationPolicyFields[]
    totalCount: number | null
    pageInfo: { endCursor: string | null; hasNextPage: boolean }
}

const POLICIES_CONFIGURATION = gql`
    query CodeIntelligenceConfigurationPolicies(
        $repository: ID
        $query: String
        $forDataRetention: Boolean
        $forIndexing: Boolean
        $first: Int
        $after: String
    ) {
        codeIntelligenceConfigurationPolicies(
            repository: $repository
            query: $query
            forDataRetention: $forDataRetention
            forIndexing: $forIndexing
            first: $first
            after: $after
        ) {
            nodes {
                ...CodeIntelligenceConfigurationPolicyFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }

    ${defaultCodeIntelligenceConfigurationPolicyFieldsFragment}
`

export const queryPolicies = (
    {
        repository,
        first,
        query,
        forDataRetention,
        forIndexing,
        after,
    }: GQL.ICodeIntelligenceConfigurationPoliciesOnQueryArguments,
    client: ApolloClient<object>
): Observable<PolicyConnection> => {
    const vars: CodeIntelligenceConfigurationPoliciesVariables = {
        repository: repository ?? null,
        query: query ?? null,
        forDataRetention: forDataRetention ?? null,
        forIndexing: forIndexing ?? null,
        first: first ?? null,
        after: after ?? null,
    }

    return from(
        client.query<CodeIntelligenceConfigurationPoliciesResult>({
            query: getDocumentNode(POLICIES_CONFIGURATION),
            variables: vars,
        })
    ).pipe(
        map(({ data }) => data),
        map(({ codeIntelligenceConfigurationPolicies }) => codeIntelligenceConfigurationPolicies)
    )
}
