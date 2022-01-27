import { gql } from '@sourcegraph/http-client'
export const FETCH_REFERENCES_QUERY = gql`
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
                                resource {
                                    path
                                    content
                                    repository {
                                        name
                                    }
                                    commit {
                                        oid
                                    }
                                }
                                range {
                                    start {
                                        line
                                        character
                                    }
                                    end {
                                        line
                                        character
                                    }
                                }
                            }
                            pageInfo {
                                endCursor
                            }
                        }
                        definitions(line: $line, character: $character) {
                            nodes {
                                resource {
                                    path
                                    content
                                    repository {
                                        name
                                    }
                                    commit {
                                        oid
                                    }
                                }
                                range {
                                    start {
                                        line
                                        character
                                    }
                                    end {
                                        line
                                        character
                                    }
                                }
                            }
                        }
                        hover(line: $line, character: $character) {
                            markdown {
                                html
                            }
                        }
                    }
                }
            }
        }
    }
`
