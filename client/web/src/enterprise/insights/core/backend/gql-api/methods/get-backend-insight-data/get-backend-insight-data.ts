import { ApolloClient } from '@apollo/client'
import { from, Observable, of, throwError } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { GetInsightViewResult, InsightViewFiltersInput } from '../../../../../../../graphql-operations'
import { BackendInsight } from '../../../../types'
import { BackendInsightData } from '../../../code-insights-backend-types'
import { InsightInProcessError } from '../../../utils/errors'
import { GET_INSIGHT_VIEW_GQL } from '../../gql/GetInsightView'

import { createBackendInsightData } from './deserializators'

export const getBackendInsightData = (
    client: ApolloClient<unknown>,
    insight: BackendInsight
): Observable<BackendInsightData> => {
    const filters: InsightViewFiltersInput = {
        includeRepoRegex: insight.filters?.includeRepoRegexp,
        excludeRepoRegex: insight.filters?.excludeRepoRegexp,
    }

    return from(
        // TODO: Use watchQuery instead of query when setting migration api is deprecated
        client.query<GetInsightViewResult>({
            query: GET_INSIGHT_VIEW_GQL,
            variables: { id: insight.id, filters },
        })
    ).pipe(
        // Note: this insight is guaranteed to exist since this function
        // is only called from within a loop of insight ids
        map(({ data }) => data.insightViews.nodes[0]),
        switchMap(data => (!data ? throwError(new InsightInProcessError()) : of(data))),
        map(data => createBackendInsightData(insight, data))
    )
}
