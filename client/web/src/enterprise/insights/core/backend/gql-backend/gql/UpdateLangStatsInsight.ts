import { gql } from '@apollo/client'

import { INSIGHT_VIEW_FRAGMENT } from './GetInsights'

export const UPDATE_LANG_STATS_INSIGHT_GQL = gql`
    mutation UpdateLangStatsInsight($id: ID!, $input: UpdatePieChartSearchInsightInput!) {
        updatePieChartSearchInsight(id: $id, input: $input) {
            view {
                ...InsightViewNode
            }
        }
    }
    ${INSIGHT_VIEW_FRAGMENT}
`
