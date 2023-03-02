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

export const ADD_CODE_HOST = gql`
    mutation AddRemoteCodeHost($input: AddExternalServiceInput!) {
        addExternalService(input: $input) {
            ...CodeHost
        }
    }

    ${CODE_HOST_FRAGMENT}
`

export const UPDATE_CODE_HOST = gql`
    mutation UpdateRemoteCodeHost($input: UpdateExternalServiceInput!) {
        updateExternalService(input: $input) {
            ...CodeHost
        }
    }

    ${CODE_HOST_FRAGMENT}
`

export const DELETE_CODE_HOST = gql`
    mutation DeleteCodeHost($id: ID!) {
        deleteExternalService(externalService: $id) {
            alwaysNil
        }
    }
`
export const GET_REPOSITORIES_BY_SERVICE = gql`
    query Repositories($first: Int, $externalService: ID) {
        repositories(first: $first, externalService: $externalService) {
            nodes {
                __typename
                id
                name
                description
            }
        }
    }
`
