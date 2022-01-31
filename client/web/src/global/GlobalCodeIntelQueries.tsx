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
        highlight(disableTimeout: false) {
            aborted
            html
        }
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
    ) {
        repository(name: $repository) {
            commit(rev: $commit) {
                blob(path: $path) {
                    lsif {
                        references(line: $line, character: $character, after: $after) {
                            nodes {
                                ...LocationFields
                            }
                            pageInfo {
                                endCursor
                            }
                        }
                        definitions(line: $line, character: $character) {
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
