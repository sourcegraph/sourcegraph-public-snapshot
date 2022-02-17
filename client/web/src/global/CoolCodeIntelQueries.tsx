import { gql } from '@sourcegraph/http-client'

const codeIntelLocationsFragments = gql`
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

const codeIntelHoverFragment = gql`
    fragment HoverFields on Hover {
        markdown {
            html
            text
        }
    }
`

export const FETCH_ALL_CODE_INTEL_QUERY = gql`
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
            commit(rev: $commit) {
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

    ${codeIntelLocationsFragments}
    ${codeIntelHoverFragment}
`

export const FETCH_REFERENCES_QUERY = gql`
    query CoolCodeIntelMoreReferences(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $first: Int
        $after: String
        $filter: String
    ) {
        repository(name: $repository) {
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        references(line: $line, character: $character, first: $first, after: $after, filter: $filter) {
                            ...LocationConnectionFields
                        }
                    }
                }
            }
        }
    }

    ${codeIntelLocationsFragments}
`

export const FETCH_DEFINITIONS_QUERY = gql`
    query CoolCodeIntelMoreDefinitions(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $filter: String
    ) {
        repository(name: $repository) {
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        definitions(line: $line, character: $character, filter: $filter) {
                            ...LocationConnectionFields
                        }
                    }
                }
            }
        }
    }

    ${codeIntelLocationsFragments}
`

export const FETCH_IMPLEMENTATIONS_QUERY = gql`
    query CoolCodeIntelMoreImplementations(
        $repository: String!
        $commit: String!
        $path: String!
        $line: Int!
        $character: Int!
        $filter: String
        $after: String
    ) {
        repository(name: $repository) {
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        implementations(line: $line, character: $character, after: $after, filter: $filter) {
                            ...LocationConnectionFields
                        }
                    }
                }
            }
        }
    }

    ${codeIntelLocationsFragments}
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
