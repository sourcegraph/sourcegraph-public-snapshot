import { gql } from '@sourcegraph/http-client'

const codeIntelFragments = gql`
    fragment LocationFields on Location {
        url
        resource {
            ...GitBlobFields
        }
        range {
            ...RangeFields
        }
    }

    fragment LocationConnectionFields on LocationConnection {
        nodes {
            ...LocationFields
        }
        pageInfo {
            endCursor
        }
    }

    fragment GitBlobFields on GitBlob {
        path
        content
        repository {
            name
        }
        commit {
            oid
        }
    }

    fragment RangeFields on Range {
        start {
            line
            character
        }
        end {
            line
            character
        }
    }
`

const gitBlobLsifDataQueryFragment = gql`
    fragment PreciseCodeIntelForLocationFields on GitBlobLSIFData {
        references(
            line: $line
            character: $character
            first: $firstReferences
            after: $afterReferences
            filter: $filter
        ) {
            ...LocationConnectionFields
        }
        implementations(
            line: $line
            character: $character
            first: $firstImplementations
            after: $afterImplementations
            filter: $filter
        ) {
            ...LocationConnectionFields
        }
        definitions(line: $line, character: $character, filter: $filter) {
            ...LocationConnectionFields
        }
    }
`

export const USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY = gql`
    query UsePreciseCodeIntelForPosition(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $afterReferences: String
        $firstReferences: Int
        $afterImplementations: String
        $firstImplementations: Int
        $filter: String
    ) {
        repository(name: $repository) {
            id
            commit(rev: $commit) {
                id
                blob(path: $path) {
                    lsif {
                        ...PreciseCodeIntelForLocationFields
                    }
                }
            }
        }
    }

    ${codeIntelFragments}
    ${gitBlobLsifDataQueryFragment}
`

export const LOAD_ADDITIONAL_REFERENCES_QUERY = gql`
    query LoadAdditionalReferences(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $afterReferences: String
        $firstReferences: Int
        $filter: String
    ) {
        repository(name: $repository) {
            id
            commit(rev: $commit) {
                id
                blob(path: $path) {
                    lsif {
                        references(
                            line: $line
                            character: $character
                            first: $firstReferences
                            after: $afterReferences
                            filter: $filter
                        ) {
                            ...LocationConnectionFields
                        }
                    }
                }
            }
        }
    }

    ${codeIntelFragments}
`

export const LOAD_ADDITIONAL_IMPLEMENTATIONS_QUERY = gql`
    query LoadAdditionalImplementations(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $afterImplementations: String
        $firstImplementations: Int
        $filter: String
    ) {
        repository(name: $repository) {
            id
            commit(rev: $commit) {
                id
                blob(path: $path) {
                    lsif {
                        implementations(
                            line: $line
                            character: $character
                            first: $firstImplementations
                            after: $afterImplementations
                            filter: $filter
                        ) {
                            ...LocationConnectionFields
                        }
                    }
                }
            }
        }
    }

    ${codeIntelFragments}
`

export const FETCH_HIGHLIGHTED_BLOB = gql`
    fragment HighlightedGitBlobFields on GitBlob {
        highlight(disableTimeout: false) {
            aborted
            html
        }
    }

    query CoolCodeIntelHighlightedBlob($repository: String!, $commit: String!, $path: String!) {
        repository(name: $repository) {
            id
            commit(rev: $commit) {
                id
                blob(path: $path) {
                    ...HighlightedGitBlobFields
                }
            }
        }
    }
`
