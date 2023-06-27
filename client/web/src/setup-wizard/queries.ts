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
        lastSyncError
        config
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

export const ADD_LOCAL_REPOSITORIES = gql`
    mutation AddLocalRepositories($paths: [String!]!) {
        addLocalRepositories(paths: $paths) {
            __typename
        }
    }
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
    mutation DeleteRemoteCodeHost($id: ID!) {
        deleteExternalService(externalService: $id) {
            alwaysNil
        }
    }
`
