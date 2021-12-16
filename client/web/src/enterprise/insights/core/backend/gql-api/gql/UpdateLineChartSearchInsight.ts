import { gql } from '@apollo/client'

import { INSIGHT_VIEW_FRAGMENT } from './GetInsights'

export const UPDATE_LINE_CHART_SEARCH_INSIGHT_GQL = gql`
    mutation UpdateLineChartSearchInsight($input: UpdateLineChartSearchInsightInput!, $id: ID!) {
        updateLineChartSearchInsight(input: $input, id: $id) {
            view {
                ...InsightViewNode
            }
        }
    }
    ${INSIGHT_VIEW_FRAGMENT}
`
