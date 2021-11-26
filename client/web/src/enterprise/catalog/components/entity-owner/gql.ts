import { useQuery, gql } from '@sourcegraph/http-client'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { GROUP_LINK_FRAGMENT } from '../../pages/group/gql2'

export const COMPONENT_OWNER_FRAGMENT = gql`
    fragment ComponentOwnerFields on Component {
        owner {
            __typename
            ... on Person {
                ...PersonLinkFields
                avatarURL
            }
            ... on Group {
                ...GroupLinkFields
                members {
                    ...PersonLinkFields
                    avatarURL
                }
                ancestorGroups {
                    ...GroupLinkFields
                }
            }
        }
    }
    ${personLinkFieldsFragment}
    ${GROUP_LINK_FRAGMENT}
`
