import { useEffect, useState } from 'react'

import { gql, useApolloClient } from '@apollo/client'

import { IsCodeInsightsLicensedResult } from '../../../graphql-operations'
import { CodeInsightsBackend, CodeInsightsGqlBackend, CodeInsightsGqlBackendLimited } from '../core'

/**
 * Returns the full or limited version of the API based on
 * whether Code Insights is licensed
 */
export function useApi(): CodeInsightsBackend | null {
    const apolloClient = useApolloClient()
    const [api, setApi] = useState<CodeInsightsBackend | null>(null)

    useEffect(() => {
        apolloClient
            .query<IsCodeInsightsLicensedResult>({
                query: gql`
                    query IsCodeInsightsLicensed {
                        enterpriseLicenseHasFeature(feature: "code-insights")
                    }
                `,
            })
            .then(result => {
                const licened = result.data.enterpriseLicenseHasFeature
                setApi(
                    licened ? new CodeInsightsGqlBackend(apolloClient) : new CodeInsightsGqlBackendLimited(apolloClient)
                )
            })
            .catch(() => new Error('Something went wrong fetching the license.'))
    }, [apolloClient])

    return api
}
