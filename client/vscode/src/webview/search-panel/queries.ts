import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const searchQuery = gql`
    query Search($query: String!, $patternType: SearchPatternType) {
        search(version: V2, query: $query, patternType: $patternType) {
            results {
                results {
                    __typename
                    ... on FileMatch {
                        ...FileMatchFields
                    }
                    ... on CommitSearchResult {
                        ...CommitSearchResultFields
                    }
                    ... on Repository {
                        description
                        ...RepositoryFields
                    }
                }
                ...SearchResultsAlertFields
                ...DynamicFiltersFields
            }
        }
    }
    fragment FileMatchFields on FileMatch {
        repository {
            ...RepositoryFields
        }
        file {
            name
            path
            url
            content
            commit {
                oid
            }
        }
        lineMatches {
            preview
            lineNumber
            offsetAndLengths
        }
        symbols {
            url
            name
            containerName
            kind
        }
    }
    fragment CommitSearchResultFields on CommitSearchResult {
        label {
            text
        }
        detail {
            text
        }
        commit {
            url
            message
            author {
                person {
                    name
                }
            }
            repository {
                ...RepositoryFields
            }
        }
        matches {
            url
            body {
                text
            }
            highlights {
                character
                line
                length
            }
        }
    }
    fragment RepositoryFields on Repository {
        name
        url
        stars
    }
    fragment SearchResultsAlertFields on SearchResults {
        alert {
            title
            description
            proposedQueries {
                description
                query
            }
        }
    }
    fragment DynamicFiltersFields on SearchResults {
        dynamicFilters {
            value
            label
            count
            limitHit
            kind
        }
    }
`
// TODO: suggestions
