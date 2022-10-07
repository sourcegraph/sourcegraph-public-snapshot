import { useApolloClient } from '@apollo/client'

import { CodeInsightsBackend, CodeInsightsGqlBackend } from '../core'

/**
 * Returns the full or limited version of the API based on
 * whether Code Insights is licensed
 */
export function useApi(): CodeInsightsBackend {
    const apolloClient = useApolloClient()

    return new CodeInsightsGqlBackend(apolloClient)
}
