import { gql } from '@sourcegraph/http-client'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'

export const COMPONENT_CODE_OWNERS_FRAGMENT = gql`
    fragment ComponentCodeOwnersFields on Component {
        codeOwners {
            edges {
                node {
                    ...PersonLinkFields
                    avatarURL
                }
                fileCount
                fileProportion
            }
        }
    }
    ${personLinkFieldsFragment}
`

export const COMPONENT_COMMITS_FRAGMENT = gql`
    fragment ComponentCommitsFields on Component {
        commits(first: 10) {
            nodes {
                ...GitCommitFields
            }
        }
    }
    ${gitCommitFragment}
`

export const COMPONENT_AUTHORS_FRAGMENT = gql`
    fragment ComponentAuthorsFields on Component {
        contributors {
            edges {
                person {
                    ...PersonLinkFields
                    avatarURL
                }
                authoredLineCount
                authoredLineProportion
                lastCommit {
                    author {
                        date
                    }
                }
            }
        }
    }
    ${personLinkFieldsFragment}
`

export const COMPONENT_USAGE_PEOPLE_FRAGMENT = gql`
    fragment ComponentUsagePeopleFields on Component {
        usage {
            people {
                node {
                    ...PersonLinkFields
                    avatarURL
                }
                authoredLineCount
                lastCommit {
                    author {
                        date
                    }
                }
            }
        }
    }
`
