import { gql } from '@sourcegraph/http-client'

const codeIntelFragments = gql`
    fragment LocationFields on Location {
        resource {
            ...GitBlobFields
        }
        range {
            ...RangeFields
        }
        url
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

const hoverFragments = gql`
    fragment HoverFields on Hover {
        markdown {
            html
            text
        }
    }
`
export const FETCH_REFERENCES_QUERY = gql`
    query CoolCodeIntelReferences(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $after: String
        $filter: String
    ) {
        repository(name: $repository) {
            __typename
            id
            commit(rev: $commit) {
                __typename
                id
                blob(path: $path) {
                    lsif {
                        references(line: $line, character: $character, after: $after, filter: $filter) {
                            ...LocationConnectionFields
                        }
                        implementations(line: $line, character: $character, after: $after, filter: $filter) {
                            ...LocationConnectionFields
                        }
                        definitions(line: $line, character: $character, filter: $filter) {
                            ...LocationConnectionFields
                        }
                        hover(line: $line, character: $character) {
                            ...HoverFields
                        }
                    }
                }
            }
        }
    }

    ${codeIntelFragments}
    ${hoverFragments}
`

const gitBlobLsifDataQueryFragment = gql`
    fragment RefPanelLsifDataFields on GitBlobLSIFData {
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
        hover(line: $line, character: $character) {
            ...HoverFields
        }
    }
`

export const USE_CODE_INTEL_QUERY = gql`
    query GetPreciseCodeIntel(
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
            __typename
            id
            commit(rev: $commit) {
                __typename
                id
                blob(path: $path) {
                    lsif {
                        ...RefPanelLsifDataFields
                    }
                }
            }
        }
    }

    ${codeIntelFragments}
    ${hoverFragments}
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
            __typename
            id
            commit(rev: $commit) {
                __typename
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
            __typename
            id
            commit(rev: $commit) {
                __typename
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
            commit(rev: $commit) {
                blob(path: $path) {
                    ...HighlightedGitBlobFields
                }
            }
        }
    }
`
