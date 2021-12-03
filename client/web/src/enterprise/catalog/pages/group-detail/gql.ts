import { gql } from '@sourcegraph/shared/src/graphql/graphql'

export const GROUP_DETAIL_FRAGMENT = gql`
    fragment GroupDetailFields on Group {
        __typename
        id
        name
        title
        description
        url
    }
`

export const GROUP_BY_NAME = gql`
    query GroupByName($name: String!) {
        group(name: $name) {
            ...GroupDetailFields
        }
    }

    ${GROUP_DETAIL_FRAGMENT}
`
