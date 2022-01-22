import { gql } from '@sourcegraph/http-client'

import { COMPONENT_OWNER_FRAGMENT } from '../../entity-owner/gql'

export const COMPONENT_LIST_FRAGMENT = gql`
    fragment ComponentListFields on Component {
        __typename
        id
        name
        kind
        description
        lifecycle
        url
        catalogURL
        ...ComponentOwnerFields
        ...ComponentLastCommitFields
    }
    ${COMPONENT_OWNER_FRAGMENT}
    fragment ComponentLastCommitFields on Component {
        commits(first: 1) {
            nodes {
                author {
                    date
                }
            }
        }
    }
`
