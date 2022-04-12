import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import { IndexConfigurationResult } from '../../../../graphql-operations'

interface UseRepositoryConfigResult {
    configuration: string
    loadingRepository: boolean
    repositoryError: ApolloError | undefined
}

export const REPOSITORY_CONFIGURATION = gql`
    query IndexConfiguration($id: ID!) {
        node(id: $id) {
            ...RepositoryIndexConfigurationFields
        }
    }

    fragment RepositoryIndexConfigurationFields on Repository {
        __typename
        indexConfiguration {
            configuration
        }
    }
`

export const useRepositoryConfig = (id: string): UseRepositoryConfigResult => {
    const { data, loading, error } = useQuery<IndexConfigurationResult>(REPOSITORY_CONFIGURATION, {
        variables: { id },
    })

    const configuration = (data?.node?.__typename === 'Repository' && data.node.indexConfiguration?.configuration) || ''

    return {
        configuration,
        loadingRepository: loading,
        repositoryError: error,
    }
}
