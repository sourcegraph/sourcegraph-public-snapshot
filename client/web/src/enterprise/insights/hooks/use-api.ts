import { useMemo } from 'react'

import { useApolloClient } from '@apollo/client'

import { type CodeInsightsBackend, CodeInsightsGqlBackend } from '../core'

/**
 * Returns the full or limited version of the API based on
 * whether Code Insights is licensed
 */
export function useApi(): CodeInsightsBackend {
    const apolloClient = useApolloClient()

    return useMemo(() => new CodeInsightsGqlBackend(apolloClient), [apolloClient])
}
