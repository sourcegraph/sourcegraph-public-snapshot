import { gql } from '@apollo/client'

export const CODE_HOST_FRAGMENT = gql`
    fragment CodeHost on ExternalService {
        __typename
        id
        kind
        repoCount
        displayName
        lastSyncAt
        nextSyncAt
    }
`

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

export const ADD_CODE_HOST = gql`
    mutation AddRemoteCodeHost($input: AddExternalServiceInput!) {
        addExternalService(input: $input) {
            ...CodeHost
        }
    }

    ${CODE_HOST_FRAGMENT}
`
