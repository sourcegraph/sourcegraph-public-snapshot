import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../../backend/graphql'
import { AnalyticsDateRange, SearchStatisticsResult, SearchStatisticsVariables } from '../../graphql-operations'

const analyticsStatItemFragment = gql`
    fragment AnalyticsStatItemFragment on AnalyticsStatItem {
        nodes {
            date
            count
            uniqueUsers
        }
        summary {
            totalCount
            totalUniqueUsers
        }
    }
`

export const fetchSearchStatistics = (
    dateRange: AnalyticsDateRange
): Observable<SearchStatisticsResult['site']['analytics']['search']> =>
    requestGraphQL<SearchStatisticsResult, SearchStatisticsVariables>(
        gql`
            query SearchStatistics($dateRange: AnalyticsDateRange!) {
                site {
                    analytics {
                        search(dateRange: $dateRange) {
                            searches {
                                ...AnalyticsStatItemFragment
                            }
                            fileViews {
                                ...AnalyticsStatItemFragment
                            }
                            fileOpens {
                                ...AnalyticsStatItemFragment
                            }
                        }
                    }
                }
            }
            ${analyticsStatItemFragment}
        `,
        {
            dateRange,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.analytics.search)
    )
