import { gql } from '@sourcegraph/http-client'

import type { SearchPageQueryResult, SearchPageQueryVariables } from '../../../graphql-operations'
import { createPreloadedQuery, type QueryReference } from '../../backend/route-loader'

/**
 * TODO: Move everything from this file into the `SearchPage` module once the `lazy` property
 * is introduced on the route object: https://github.com/remix-run/react-router/pull/10045.
 * It lives in a separate file now to make the lazy loading of the `SearchPage` module
 * possible.
 *
 * TODO: Create this GraphQL query combining fragments specified in underliying components
 * instead of hardcoding it here. This will be possible after the previous comment is resolved.
 *
 * TODO: We need the `evaluateFeatureFlags` query that can evaluate multiple feature flags.
 */
const SEARCH_PAGE_QUERY = gql`
    query SearchPageQuery {
        externalServices {
            totalCount
        }
        codehostWidgetFlag: evaluateFeatureFlag(flagName: "plg-enable-add-codehost-widget")
    }
`

export const { queryLoader, usePreloadedQueryData } = createPreloadedQuery<
    SearchPageQueryResult,
    SearchPageQueryVariables
>(SEARCH_PAGE_QUERY)

export function loader(): Promise<Record<string, QueryReference | undefined>> {
    return queryLoader({})
}
