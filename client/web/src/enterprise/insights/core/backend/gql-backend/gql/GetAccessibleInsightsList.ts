import { gql } from '@apollo/client'

export const GET_ACCESSIBLE_INSIGHTS_LIST = gql`
    query GetAccessibleInsightsList {
        insightViews {
            nodes {
                id
                presentation {
                    __typename
                    ... on LineChartInsightViewPresentation {
                        title
                    }

                    ... on PieChartInsightViewPresentation {
                        title
                    }
                }
            }
        }
    }
`
