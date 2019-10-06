import { gql } from '../../../../../shared/src/graphql/graphql'
import { ActorQuery } from '../../../actor/graphql'

export const ThreadFragment = gql`
    fragment ThreadFragment on Thread {
        __typename
        id
        number
        title
        isDraft
        isPendingExternalCreation
        state
        kind
        url
        createdAt
        externalURLs {
            url
            serviceType
        }
        repository {
            name
            url
        }
        author {
            ${ActorQuery}
        }
        assignees {
            nodes {
                ${ActorQuery}
            }
            totalCount
        }
        comments {
            totalCount
        }
    }
`
