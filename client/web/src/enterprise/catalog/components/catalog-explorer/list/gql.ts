import { gql } from '@sourcegraph/http-client'

import { COMPONENT_OWNER_LINK_FRAGMENT } from '../../component-owner-link/ComponentOwnerLink'

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
        owner {
            ...ComponentOwnerLinkFields
        }

        commitsForLastCommitDate: commits(first: 1) {
            nodes {
                author {
                    date
                }
            }
        }
    }
    ${COMPONENT_OWNER_LINK_FRAGMENT}
`
