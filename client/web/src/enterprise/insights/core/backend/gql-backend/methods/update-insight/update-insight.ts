import { ApolloClient } from '@apollo/client'
import { ApolloCache } from '@apollo/client/cache'
import { MutationUpdaterFunction } from '@apollo/client/core/types'
import { from, Observable } from 'rxjs'

import {
    UpdateLangStatsInsightResult,
    UpdateLangStatsInsightVariables,
    UpdateLineChartSearchInsightResult,
    UpdateLineChartSearchInsightVariables,
} from '../../../../../../../graphql-operations'
import { InsightType } from '../../../../types'
import { InsightUpdateInput } from '../../../code-insights-backend-types'
import { UPDATE_LANG_STATS_INSIGHT_GQL } from '../../gql/UpdateLangStatsInsight'
import { UPDATE_LINE_CHART_SEARCH_INSIGHT_GQL } from '../../gql/UpdateLineChartSearchInsight'

import {
    getCaptureGroupInsightUpdateInput,
    getLangStatsInsightUpdateInput,
    getSearchInsightUpdateInput,
} from './serializators'

type UpdateVariables = UpdateLineChartSearchInsightVariables | UpdateLangStatsInsightVariables
export type UpdateResult = UpdateLineChartSearchInsightResult | UpdateLangStatsInsightResult

export const updateInsight = (
    client: ApolloClient<unknown>,
    input: InsightUpdateInput,
    update?: MutationUpdaterFunction<UpdateResult, UpdateVariables, unknown, ApolloCache<unknown>>
): Observable<unknown> => {
    const { nextInsightData, insightId } = input

    switch (nextInsightData.type) {
        case InsightType.SearchBased: {
            return from(
                client.mutate<UpdateLineChartSearchInsightResult, UpdateLineChartSearchInsightVariables>({
                    mutation: UPDATE_LINE_CHART_SEARCH_INSIGHT_GQL,
                    variables: { input: getSearchInsightUpdateInput(nextInsightData), id: insightId },
                    update,
                })
            )
        }

        case InsightType.CaptureGroup: {
            return from(
                client.mutate<UpdateLineChartSearchInsightResult, UpdateLineChartSearchInsightVariables>({
                    mutation: UPDATE_LINE_CHART_SEARCH_INSIGHT_GQL,
                    variables: { input: getCaptureGroupInsightUpdateInput(nextInsightData), id: insightId },
                    update,
                })
            )
        }

        case InsightType.LangStats: {
            return from(
                client.mutate<UpdateLangStatsInsightResult, UpdateLangStatsInsightVariables>({
                    mutation: UPDATE_LANG_STATS_INSIGHT_GQL,
                    variables: {
                        id: insightId,
                        input: getLangStatsInsightUpdateInput(nextInsightData),
                    },
                    update,
                })
            )
        }
    }
}
