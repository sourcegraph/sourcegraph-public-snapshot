import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { GROUP_LINK_FRAGMENT } from '../../pages/group-detail/gql2'

export const CATALOG_ENTITY_OWNER_FRAGMENT = gql`
    fragment CatalogEntityOwnerFields on CatalogEntity {
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
