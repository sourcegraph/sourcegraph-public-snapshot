import { gql } from '@apollo/client'

import { INSIGHT_VIEW_FRAGMENT } from '../backend/gql-backend/gql/GetInsights'

export const SAVE_INSIGHT_AS_NEW_VIEW_GQL = gql`
    mutation SaveInsightAsNewView($input: InsightNewViewInput!) {
        saveInsightAsNewView(input: $input) {
            view {
                ...InsightViewNode
            }
        }
    }
    ${INSIGHT_VIEW_FRAGMENT}
`

interface insightNewViewInputProps {
    insightViewId: string
    dashboardId: string?
    title: string
    filters: InsightFilters
}

export function useSaveInsightAsNewView(props: insightNewViewInputProps) {
    const { insightViewId, dashboardId } = props

}
