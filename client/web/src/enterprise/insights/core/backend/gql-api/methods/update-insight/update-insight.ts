import { ApolloClient } from '@apollo/client'
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

export const updateInsight = (client: ApolloClient<unknown>, input: InsightUpdateInput): Observable<unknown> => {
    const insight = input.newInsight
    const oldInsight = input.oldInsight

    switch (insight.viewType) {
        case InsightType.SearchBased: {
            return from(
                client.mutate<UpdateLineChartSearchInsightResult, UpdateLineChartSearchInsightVariables>({
                    mutation: UPDATE_LINE_CHART_SEARCH_INSIGHT_GQL,
                    variables: { input: getSearchInsightUpdateInput(insight), id: oldInsight.id },
                })
            )
        }

        case InsightType.CaptureGroup: {
            return from(
                client.mutate<UpdateLineChartSearchInsightResult, UpdateLineChartSearchInsightVariables>({
                    mutation: UPDATE_LINE_CHART_SEARCH_INSIGHT_GQL,
                    variables: { input: getCaptureGroupInsightUpdateInput(insight), id: oldInsight.id },
                })
            )
        }

        case InsightType.LangStats: {
            return from(
                client.mutate<UpdateLangStatsInsightResult, UpdateLangStatsInsightVariables>({
                    mutation: UPDATE_LANG_STATS_INSIGHT_GQL,
                    variables: {
                        id: oldInsight.id,
                        input: getLangStatsInsightUpdateInput(insight),
                    },
                })
            )
        }
    }
}
