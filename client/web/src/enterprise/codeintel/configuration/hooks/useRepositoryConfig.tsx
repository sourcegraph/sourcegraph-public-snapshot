import type { ApolloError } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'

import type { AutoIndexJobDescriptionFields, IndexConfigurationResult } from '../../../../graphql-operations'
import { REPOSITORY_CONFIGURATION } from '../backend'

interface UseRepositoryConfigResult {
    configuration: {
        raw: string
        parsed: AutoIndexJobDescriptionFields[]
    } | null
    loadingRepository: boolean
    repositoryError: ApolloError | undefined
}

export const useRepositoryConfig = (id: string): UseRepositoryConfigResult => {
    const { data, loading, error } = useQuery<IndexConfigurationResult>(REPOSITORY_CONFIGURATION, {
        variables: { id },
    })

    const configuration =
        data?.node?.__typename === 'Repository' && data.node.indexConfiguration?.configuration
            ? {
                  raw: data.node.indexConfiguration.configuration,
                  parsed: data.node.indexConfiguration.parsedConfiguration ?? [],
              }
            : null

    return {
        configuration,
        loadingRepository: loading,
        repositoryError: error,
    }
}
