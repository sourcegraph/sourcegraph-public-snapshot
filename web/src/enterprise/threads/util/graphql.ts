import { gql } from '../../../../../shared/src/graphql/graphql'

export const ThreadFragment = gql`
    fragment ThreadFragment on Thread {
        __typename
        id
        number
        title
        state
        kind
        url
        externalURLs {
            url
            serviceType
        }
        repository {
            name
        }
    }
`
