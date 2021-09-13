import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import {
    createInvalidGraphQLMutationResponseError,
    dataOrThrowErrors,
    gql,
} from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    CodeIntelligenceConfigurationPoliciesResult,
    CodeIntelligenceConfigurationPoliciesVariables,
    CodeIntelligenceConfigurationPolicyFields,
    CodeIntelligenceConfigurationPolicyResult,
    CodeIntelligenceConfigurationPolicyVariables,
    CreateCodeIntelligenceConfigurationPolicyResult,
    CreateCodeIntelligenceConfigurationPolicyVariables,
    DeleteCodeIntelligenceConfigurationPolicyResult,
    DeleteCodeIntelligenceConfigurationPolicyVariables,
    IndexConfigurationResult,
    IndexConfigurationVariables,
    InferredIndexConfigurationResult,
    InferredIndexConfigurationVariables,
    RepositoryBranchesFields,
    RepositoryIndexConfigurationFields,
    RepositoryInferredIndexConfigurationFields,
    RepositoryNameFields,
    RepositoryNameResult,
    RepositoryNameVariables,
    RepositoryTagsFields,
    SearchGitBranchesResult,
    SearchGitBranchesVariables,
    SearchGitTagsResult,
    SearchGitTagsVariables,
    UpdateCodeIntelligenceConfigurationPolicyResult,
    UpdateCodeIntelligenceConfigurationPolicyVariables,
    UpdateRepositoryIndexConfigurationResult,
    UpdateRepositoryIndexConfigurationVariables,
} from '../../../graphql-operations'

const codeIntelligenceConfigurationPolicyFieldsFragment = gql`
    fragment CodeIntelligenceConfigurationPolicyFields on CodeIntelligenceConfigurationPolicy {
        __typename
        id
        name
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

export function getPolicies(repositoryId?: string): Observable<CodeIntelligenceConfigurationPolicyFields[]> {
    const query = gql`
        query CodeIntelligenceConfigurationPolicies($repositoryId: ID) {
            codeIntelligenceConfigurationPolicies(repository: $repositoryId) {
                ...CodeIntelligenceConfigurationPolicyFields
            }
        }

        ${codeIntelligenceConfigurationPolicyFieldsFragment}
    `

    return requestGraphQL<CodeIntelligenceConfigurationPoliciesResult, CodeIntelligenceConfigurationPoliciesVariables>(
        query,
        { repositoryId: repositoryId ?? null }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ codeIntelligenceConfigurationPolicies }) => codeIntelligenceConfigurationPolicies)
    )
}

export function getPolicyById(id: string): Observable<CodeIntelligenceConfigurationPolicyFields | undefined> {
    const query = gql`
        query CodeIntelligenceConfigurationPolicy($id: ID!) {
            node(id: $id) {
                ...CodeIntelligenceConfigurationPolicyFields
            }
        }

        ${codeIntelligenceConfigurationPolicyFieldsFragment}
    `

    return requestGraphQL<CodeIntelligenceConfigurationPolicyResult, CodeIntelligenceConfigurationPolicyVariables>(
        query,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such CodeIntelligenceConfigurationPolicy')
            }
            return node
        })
    )
}

export function updatePolicy(
    policy: CodeIntelligenceConfigurationPolicyFields,
    repositoryId?: string
): Observable<void> {
    if (policy.id) {
        const query = gql`
            mutation UpdateCodeIntelligenceConfigurationPolicy(
                $id: ID!
                $name: String!
                $type: GitObjectType!
                $pattern: String!
                $retentionEnabled: Boolean!
                $retentionDurationHours: Int
                $retainIntermediateCommits: Boolean!
                $indexingEnabled: Boolean!
                $indexCommitMaxAgeHours: Int
                $indexIntermediateCommits: Boolean!
            ) {
                updateCodeIntelligenceConfigurationPolicy(
                    id: $id
                    name: $name
                    type: $type
                    pattern: $pattern
                    retentionEnabled: $retentionEnabled
                    retentionDurationHours: $retentionDurationHours
                    retainIntermediateCommits: $retainIntermediateCommits
                    indexingEnabled: $indexingEnabled
                    indexCommitMaxAgeHours: $indexCommitMaxAgeHours
                    indexIntermediateCommits: $indexIntermediateCommits
                ) {
                    alwaysNil
                }
            }
        `

        return requestGraphQL<
            UpdateCodeIntelligenceConfigurationPolicyResult,
            UpdateCodeIntelligenceConfigurationPolicyVariables
        >(query, { ...policy }).pipe(
            map(dataOrThrowErrors),
            map(() => {
                // no-op
            })
        )
    }

    const query = gql`
        mutation CreateCodeIntelligenceConfigurationPolicy(
            $repositoryId: ID
            $name: String!
            $type: GitObjectType!
            $pattern: String!
            $retentionEnabled: Boolean!
            $retentionDurationHours: Int
            $retainIntermediateCommits: Boolean!
            $indexingEnabled: Boolean!
            $indexCommitMaxAgeHours: Int
            $indexIntermediateCommits: Boolean!
        ) {
            createCodeIntelligenceConfigurationPolicy(
                repository: $repositoryId
                name: $name
                type: $type
                pattern: $pattern
                retentionEnabled: $retentionEnabled
                retentionDurationHours: $retentionDurationHours
                retainIntermediateCommits: $retainIntermediateCommits
                indexingEnabled: $indexingEnabled
                indexCommitMaxAgeHours: $indexCommitMaxAgeHours
                indexIntermediateCommits: $indexIntermediateCommits
            ) {
                id
            }
        }
    `

    return requestGraphQL<
        CreateCodeIntelligenceConfigurationPolicyResult,
        CreateCodeIntelligenceConfigurationPolicyVariables
    >(query, { ...policy, repositoryId: repositoryId ?? null }).pipe(
        map(dataOrThrowErrors),
        map(() => {
            // no-op
        })
    )
}

export function deletePolicyById(id: string): Observable<void> {
    const query = gql`
        mutation DeleteCodeIntelligenceConfigurationPolicy($id: ID!) {
            deleteCodeIntelligenceConfigurationPolicy(policy: $id) {
                alwaysNil
            }
        }
    `

    return requestGraphQL<
        DeleteCodeIntelligenceConfigurationPolicyResult,
        DeleteCodeIntelligenceConfigurationPolicyVariables
    >(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(() => {
            // no-op
        })
    )
}

export function getConfigurationForRepository(id: string): Observable<RepositoryIndexConfigurationFields | null> {
    const query = gql`
        query IndexConfiguration($id: ID!) {
            node(id: $id) {
                ...RepositoryIndexConfigurationFields
            }
        }

        fragment RepositoryIndexConfigurationFields on Repository {
            __typename
            indexConfiguration {
                configuration
            }
        }
    `

    return requestGraphQL<IndexConfigurationResult, IndexConfigurationVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such Repository')
            }
            return node
        })
    )
}

export function getInferredConfigurationForRepository(
    id: string
): Observable<RepositoryInferredIndexConfigurationFields | null> {
    const query = gql`
        query InferredIndexConfiguration($id: ID!) {
            node(id: $id) {
                ...RepositoryInferredIndexConfigurationFields
            }
        }

        fragment RepositoryInferredIndexConfigurationFields on Repository {
            __typename
            indexConfiguration {
                inferredConfiguration
            }
        }
    `

    return requestGraphQL<InferredIndexConfigurationResult, InferredIndexConfigurationVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such Repository')
            }
            return node
        })
    )
}

export function updateConfigurationForRepository(id: string, content: string): Observable<void> {
    const query = gql`
        mutation UpdateRepositoryIndexConfiguration($id: ID!, $content: String!) {
            updateRepositoryIndexConfiguration(repository: $id, configuration: $content) {
                alwaysNil
            }
        }
    `

    return requestGraphQL<UpdateRepositoryIndexConfigurationResult, UpdateRepositoryIndexConfigurationVariables>(
        query,
        {
            id,
            content,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.updateRepositoryIndexConfiguration) {
                throw createInvalidGraphQLMutationResponseError('UpdateRepositoryIndexConfiguration')
            }
        })
    )
}

export function searchGitTags(id: string, term: string): Observable<RepositoryTagsFields | null> {
    const query = gql`
        query SearchGitTags($id: ID!, $query: String!) {
            node(id: $id) {
                ...RepositoryTagsFields
            }
        }

        fragment RepositoryTagsFields on Repository {
            __typename
            name
            tags(query: $query, first: 10) {
                nodes {
                    displayName
                }

                totalCount
            }
        }
    `

    return requestGraphQL<SearchGitTagsResult, SearchGitTagsVariables>(query, { id, query: term }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such Repository')
            }
            return node
        })
    )
}

export function searchGitBranches(id: string, term: string): Observable<RepositoryBranchesFields | null> {
    const query = gql`
        query SearchGitBranches($id: ID!, $query: String!) {
            node(id: $id) {
                ...RepositoryBranchesFields
            }
        }

        fragment RepositoryBranchesFields on Repository {
            __typename
            name
            branches(query: $query, first: 10) {
                nodes {
                    displayName
                }

                totalCount
            }
        }
    `

    return requestGraphQL<SearchGitBranchesResult, SearchGitBranchesVariables>(query, { id, query: term }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such Repository')
            }
            return node
        })
    )
}

export function repoName(id: string): Observable<RepositoryNameFields | null> {
    const query = gql`
        query RepositoryName($id: ID!) {
            node(id: $id) {
                ...RepositoryNameFields
            }
        }

        fragment RepositoryNameFields on Repository {
            __typename
            name
        }
    `

    return requestGraphQL<RepositoryNameResult, RepositoryNameVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such Repository')
            }
            return node
        })
    )
}
