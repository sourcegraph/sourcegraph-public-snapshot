import { gql } from '@apollo/client'

import { CODE_HOST_FRAGMENT } from '../../queries'

export const GET_CODE_HOSTS = gql`
    query GetCodeHosts {
        externalServices {
            nodes {
                ...CodeHost
            }
        }
    }

    ${CODE_HOST_FRAGMENT}
`

export const GET_CODE_HOST_BY_ID = gql`
    query GetExternalServiceById($id: ID!) {
        node(id: $id) {
            ... on ExternalService {
                id
                __typename
                kind
                displayName
                repoCount
                config
            }
        }
    }
`

export const DELETE_CODE_HOST = gql`
    mutation DeleteCodeHost($id: ID!) {
        deleteExternalService(externalService: $id) {
            alwaysNil
        }
    }
`
