import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'
import { COMPONENT_STATE_FRAGMENT } from '../overview/components/catalog-explorer/gql'

import { GROUP_LINK_FRAGMENT } from './gql2'

const GROUP_MEMBERS_FRAGMENT = gql`
    fragment GroupMembersFields on Group {
        members {
            ...PersonLinkFields
            avatarURL
        }
    }
    ${personLinkFieldsFragment}
`

const GROUP_PARENT_GROUP_FRAGMENT = gql`
    fragment GroupParentGroupFields on Group {
        parentGroup {
            ...GroupLinkFields
        }
    }
    ${GROUP_LINK_FRAGMENT}
`

const GROUP_ANCESTOR_GROUPS_FRAGMENT = gql`
    fragment GroupAncestorGroupsFields on Group {
        ancestorGroups {
            ...GroupLinkFields
        }
    }
    ${GROUP_LINK_FRAGMENT}
`

const GROUP_CHILD_GROUPS_FRAGMENT = gql`
    fragment GroupChildGroupsFields on Group {
        childGroups {
            ...GroupLinkFields
            members {
                __typename
            }
        }
    }
    ${GROUP_LINK_FRAGMENT}
`

const GROUP_OWNED_ENTITIES_FRAGMENT = gql`
    fragment GroupOwnedEntitiesFields on Group {
        ownedEntities {
            id
            name
            description
            url
            ... on Component {
                kind
            }
            ...ComponentStateFields
        }
    }
    ${COMPONENT_STATE_FRAGMENT}
`

const GROUP_DETAIL_FRAGMENT = gql`
    fragment GroupDetailFields on Group {
        __typename
        id
        name
        title
        description
        url
        ...GroupMembersFields
        ...GroupParentGroupFields
        ...GroupAncestorGroupsFields
        ...GroupChildGroupsFields
        ...GroupOwnedEntitiesFields
    }
    ${GROUP_MEMBERS_FRAGMENT}
    ${GROUP_PARENT_GROUP_FRAGMENT}
    ${GROUP_ANCESTOR_GROUPS_FRAGMENT}
    ${GROUP_CHILD_GROUPS_FRAGMENT}
    ${GROUP_OWNED_ENTITIES_FRAGMENT}
`

export const GROUP_BY_NAME = gql`
    query GroupByName($name: String!) {
        group(name: $name) {
            ...GroupDetailFields
        }
    }

    ${GROUP_DETAIL_FRAGMENT}
`
