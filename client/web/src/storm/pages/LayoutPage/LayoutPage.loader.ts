import { ApolloError } from '@apollo/client'

import { gql } from '@sourcegraph/http-client'

import type { LayoutPageQueryResult, LayoutPageQueryVariables } from '../../../graphql-operations'
import { createPreloadedQuery, type QueryReference } from '../../backend/route-loader'

export const siteFlagFieldsFragment = gql`
    fragment SiteFlagFields on Site {
        id
        needsRepositoryConfiguration
        freeUsersExceeded
        alerts {
            ...SiteFlagAlertFields
        }
        productSubscription {
            license {
                expiresAt
            }
            noLicenseWarningUserCount
        }
    }

    fragment SiteFlagAlertFields on Alert {
        type
        message
        isDismissibleWithKey
    }
`

const LAYOUT_PAGE_QUERY = gql`
    query LayoutPageQuery {
        # Preload site-flags. This is consumed by the GlobalAlerts component which is shared between
        # storm and non-storm pages, hence we're duplicating the query and rely fully on the Apollo
        # cache instead of createPreloadedQuery.
        site {
            ...SiteFlagFields
        }

        # Preload feature flags
        flag1: evaluateFeatureFlag(flagName: "contrast-compliant-syntax-highlighting")
        flag2: evaluateFeatureFlag(flagName: "cody")
        flag3: evaluateFeatureFlag(flagName: "blob-page-switch-areas-shortcuts")
    }

    ${siteFlagFieldsFragment}
`

const { queryLoader } = createPreloadedQuery<LayoutPageQueryResult, LayoutPageQueryVariables>(LAYOUT_PAGE_QUERY)

export async function loader(): Promise<Record<string, QueryReference | undefined>> {
    try {
        const loader = await queryLoader({})
        return loader
    } catch (error) {
        if (error instanceof ApolloError && (error.networkError as any)?.status === 401) {
            // When logged out, we do not use the loader data from the layout page. Instead, we need
            // to be certain that the error boundary is not triggered so we return an empty loader
            // object.
            return {}
        }
        throw error
    }
}
