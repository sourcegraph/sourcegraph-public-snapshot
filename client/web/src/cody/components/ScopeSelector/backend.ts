import { gql } from '@sourcegraph/http-client'

const REPO_FIELDS = gql`
    fragment ContextSelectorRepoFields on Repository {
        id
        name
        externalRepository {
            id
            serviceType
        }
    }
`

export const ReposSelectorSearchQuery = gql`
    query ReposSelectorSearch($query: String!) {
        repositories(query: $query, first: 15) {
            nodes {
                ...ContextSelectorRepoFields
            }
        }
    }

    ${REPO_FIELDS}
`

export const SuggestedReposQuery = gql`
    query SuggestedRepos($names: [String!]!, $numResults: Int!) {
        byName: repositories(names: $names, first: $numResults) {
            nodes {
                ...ContextSelectorRepoFields
            }
        }
        # We also grab the first $numResults embedded repos available on the site
        # to show in suggestions as a backup.
        firstN: repositories(first: $numResults, embedded: true) {
            nodes {
                ...ContextSelectorRepoFields
            }
        }
    }

    ${REPO_FIELDS}
`

export const ReposStatusQuery = gql`
    query ReposStatus($repoNames: [String!]!, $first: Int!) {
        repositories(names: $repoNames, first: $first) {
            nodes {
                ...ContextSelectorRepoFields
            }
        }
    }

    ${REPO_FIELDS}
`
