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
                            nodes {
                                ...LocationFields
                            }
                            pageInfo {
                                endCursor
                            }
                        }
                        definitions(line: $line, character: $character, filter: $filter) {
                            nodes {
                                ...LocationFields
                            }
                        }
                        hover(line: $line, character: $character) {
                            markdown {
                                html
                                text
                            }
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
