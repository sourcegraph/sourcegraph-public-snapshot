import { gql } from '@sourcegraph/shared/src/graphql/graphql'

import { personLinkFieldsFragment } from '../../../../person/PersonLink'

export const GROUP_LINK_FRAGMENT = gql`
    fragment GroupLinkFields on Group {
        id
        name
        title
        description
        url
    }
`

const GROUP_MEMBERS_FRAGMENT = gql`
    fragment GroupMembersFields on Group {
        members {
            ...PersonLinkFields
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

const GROUP_CHILD_GROUPS_FRAGMENT = gql`
    fragment GroupChildGroupsFields on Group {
        childGroups {
            ...GroupLinkFields
        }
    }
    ${GROUP_LINK_FRAGMENT}
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
        ...GroupChildGroupsFields
    }
    ${GROUP_MEMBERS_FRAGMENT}
    ${GROUP_PARENT_GROUP_FRAGMENT}
    ${GROUP_CHILD_GROUPS_FRAGMENT}
`

export const GROUP_BY_NAME = gql`
    query GroupByName($name: String!) {
        group(name: $name) {
            ...GroupDetailFields
        }
    }

    ${GROUP_DETAIL_FRAGMENT}
`
