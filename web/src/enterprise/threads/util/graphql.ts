import { gql } from '../../../../../shared/src/graphql/graphql'
import { ActorQuery } from '../../../actor/graphql'

export const ThreadFragment = gql`
    fragment ThreadFragment on Thread {
        __typename
        id
        title
        state
        kind
        createdAt
        externalURLs {
            url
            serviceType
        }
        repository {
            name
            url
        }
        assignees {
            nodes {
                ${ActorQuery}
            }
            totalCount
        }
        labels {
            nodes {
                name
                color
            }
            totalCount
        }
    }
`
