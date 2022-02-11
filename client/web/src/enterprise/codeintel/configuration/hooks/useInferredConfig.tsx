import { ApolloError } from '@apollo/client'

import { gql, useQuery } from '@sourcegraph/http-client'

import { InferredIndexConfigurationResult } from '../../../../graphql-operations'

export const INFERRED_CONFIGURATION = gql`
    query InferredIndexConfiguration($id: ID!) {
        node(id: $id) {
            ...RepositoryInferredIndexConfigurationFields
        }
    }

    fragment RepositoryInferredIndexConfigurationFields on Repository {
        __typename
        indexConfiguration {
            inferredConfiguration
        }
    }
`

interface UseInferredConfigResult {
    inferredConfiguration: string
    loadingInferred: boolean
    inferredError: ApolloError | undefined
}

export const useInferredConfig = (id: string): UseInferredConfigResult => {
    const { data, loading, error } = useQuery<InferredIndexConfigurationResult>(INFERRED_CONFIGURATION, {
        variables: { id },
    })

    const inferredConfiguration =
        (data?.node?.__typename === 'Repository' && data.node.indexConfiguration?.inferredConfiguration) || ''

    return {
        inferredConfiguration,
        loadingInferred: loading,
        inferredError: error,
    }
}
