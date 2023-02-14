import { ApolloError } from '@apollo/client'
import { GlobalCodeIntelStatusResult, GlobalCodeIntelStatusVariables } from '../../../../graphql-operations'
import { globalCodeIntelStatusQuery } from './queries'
import { useQuery } from '@sourcegraph/http-client'

export interface UseGlobalCodeIntelStatusParameters {
    variables: GlobalCodeIntelStatusVariables
}

export interface UseGlobalCodeIntelStatusResult {
    data?: GlobalCodeIntelStatusResult
    error?: ApolloError
    loading: boolean
}

export const useGlobalCodeIntelStatus = ({
    variables,
}: UseGlobalCodeIntelStatusParameters): UseGlobalCodeIntelStatusResult => {
    const { data, error, loading } = useQuery<GlobalCodeIntelStatusResult, GlobalCodeIntelStatusVariables>(
        globalCodeIntelStatusQuery,
        {
            variables,
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
        }
    )

    return { data, loading, error }
}
