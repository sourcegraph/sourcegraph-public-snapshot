import { gql } from '@sourcegraph/http-client'

export const FETCH_REFERENCES_QUERY = gql`
    fragment LocationFields on Location {
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

    fragment HoverFields on Hover {
        markdown {
            html
            text
        }
    }

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
