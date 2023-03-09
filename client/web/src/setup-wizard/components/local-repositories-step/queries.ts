import { gql } from '@apollo/client'

export const GET_LOCAL_DIRECTORY_PATH = gql`
    query GetLocalDirectoryPath {
        localDirectoryPicker {
            path
        }
    }
`

export const DISCOVER_LOCAL_REPOSITORIES = gql`
    query DiscoverLocalRepositories($dir: String!) {
        localDirectory(dir: $dir) {
            path
            repositories {
                __typename
                path
                name
            }
        }
    }
`

export const GET_REPOSITORIES_BY_SERVICE = gql`
    query RepositoriesByService($first: Int, $externalService: ID) {
        repositories(first: $first, externalService: $externalService) {
            nodes {
                __typename
                id
                name
                description
                url
            }
        }
    }
`
