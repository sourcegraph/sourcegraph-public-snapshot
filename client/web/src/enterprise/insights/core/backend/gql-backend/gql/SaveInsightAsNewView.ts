import { gql } from '@apollo/client'

import { INSIGHT_VIEW_FRAGMENT } from './GetInsights'

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
