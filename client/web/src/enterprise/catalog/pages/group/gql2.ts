import { gql } from '@sourcegraph/shared/src/graphql/graphql'

// TODO(sqs): moved here because of import cycle making it `undefined`

export const GROUP_LINK_FRAGMENT = gql`
    fragment GroupLinkFields on Group {
        id
        name
        title
        description
        url
    }
`
