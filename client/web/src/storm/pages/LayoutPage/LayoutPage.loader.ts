import { gql } from '@sourcegraph/http-client'

import { LayoutPageQueryResult, LayoutPageQueryVariables } from '../../../graphql-operations'
import { createPreloadedQuery, QueryReference } from '../../backend/route-loader'

export const siteFlagFieldsFragment = gql`
    fragment SiteFlagFields on Site {
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
            id
            ...SiteFlagFields
        }

        # Preload feature flags
        flag1: evaluateFeatureFlag(flagName: "contrast-compliant-syntax-highlighting")
        flag2: evaluateFeatureFlag(flagName: "cody")
        flag3: evaluateFeatureFlag(flagName: "search-ownership")
    }

    ${siteFlagFieldsFragment}
`

const { queryLoader } = createPreloadedQuery<LayoutPageQueryResult, LayoutPageQueryVariables>(LAYOUT_PAGE_QUERY)

export function loader(): Promise<Record<string, QueryReference | undefined>> {
    return queryLoader({})
}
