import { ApolloClient } from '@apollo/client'
import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { ALL_INSIGHTS_DASHBOARD } from '../../../../constants'
import { InsightDashboard } from '../../../../types'

import { getDashboards } from './get-dashboards'

export const getDashboardById = (
    apolloClient: ApolloClient<unknown>,
    input: { dashboardId: string | undefined }
): Observable<InsightDashboard | null> => {
    const { dashboardId } = input

    // the 'all' dashboardId is not a real dashboard so return nothing
    if (!dashboardId || dashboardId === ALL_INSIGHTS_DASHBOARD.id) {
        return of(null)
    }

    return getDashboards(apolloClient, dashboardId).pipe(
        map(dashboards => dashboards.find(({ id }) => id === dashboardId) ?? null)
    )
}
