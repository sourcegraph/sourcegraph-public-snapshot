import { ApolloClient } from '@apollo/client'
import { Observable, of, throwError } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { fromObservableQuery } from '@sourcegraph/http-client'

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
    const { excludeRepoRegexp, includeRepoRegexp, context } = insight.filters
    const filters: InsightViewFiltersInput = {
        includeRepoRegex: includeRepoRegexp,
        excludeRepoRegex: excludeRepoRegexp,
        searchContexts: [context],
    }

    return fromObservableQuery(
        client.watchQuery<GetInsightViewResult>({
            query: GET_INSIGHT_VIEW_GQL,
            variables: { id: insight.id, filters },
            // This query is set to network-only because the caching is not working correctly
            // https://github.com/sourcegraph/sourcegraph/issues/33813
            fetchPolicy: 'network-only',
        })
    ).pipe(
        // Note: this insight is guaranteed to exist since this function
        // is only called from within a loop of insight ids
        map(({ data }) => data.insightViews.nodes[0]),
        switchMap(data => (!data ? throwError(new InsightInProcessError()) : of(data))),
        map(data => createBackendInsightData(insight, data))
    )
}
