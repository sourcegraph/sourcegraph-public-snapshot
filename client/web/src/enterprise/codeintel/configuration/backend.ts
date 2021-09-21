import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    RepositoryBranchesFields,
    RepositoryNameFields,
    RepositoryNameResult,
    RepositoryNameVariables,
    RepositoryTagsFields,
    SearchGitBranchesResult,
    SearchGitBranchesVariables,
    SearchGitTagsResult,
    SearchGitTagsVariables,
} from '../../../graphql-operations'

export const codeIntelligenceConfigurationPolicyFieldsFragment = gql`
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
