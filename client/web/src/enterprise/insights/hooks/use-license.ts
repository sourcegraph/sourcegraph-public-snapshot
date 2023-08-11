import { useEffect, useState } from 'react'

import { gql, useApolloClient } from '@apollo/client'

import type { IsCodeInsightsLicensedResult } from '../../../graphql-operations'
import { useCodeInsightsLicenseState } from '../stores'

/**
 * Returns the full or limited version of the API based on
 * whether Code Insights is licensed
 */
export function useLicense(): boolean {
    const apolloClient = useApolloClient()
    const [fetched, setFetched] = useState<boolean>(false)

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
                const licensed = result.data.enterpriseLicenseHasFeature
                useCodeInsightsLicenseState.setState({ licensed, insightsLimit: licensed ? null : 2 })
                setFetched(true)
            })
            .catch(() => new Error('Something went wrong fetching the license.'))
    }, [apolloClient])

    return fetched
}
