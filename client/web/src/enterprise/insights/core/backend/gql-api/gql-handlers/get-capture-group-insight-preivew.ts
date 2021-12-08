import { ApolloClient, gql } from '@apollo/client'
import { LineChartContent } from 'sourcegraph'

import {
    GetCaptureGroupInsightPreviewResult,
    GetCaptureGroupInsightPreviewVariables,
} from '../../../../../../graphql-operations'
import { LINE_CHART_WITH_HUGE_NUMBER_OF_LINES } from '../../../../../../views/mocks/charts-content'
import { CaptureInsightSettings } from '../../code-insights-backend-types'
import { getStepInterval } from '../utils/insight-transformers'

const GET_CAPTURE_GROUP_INSIGHT_PREVIEW_GQL = gql`
    query GetCaptureGroupInsightPreview($input: SearchInsightLivePreviewInput!) {
        searchInsightLivePreview(input: $input) {
            points {
                dateTime
                value
            }
            label
        }
    }
`

export const getCaptureGroupInsightsPreview = (
    client: ApolloClient<unknown>,
    input: CaptureInsightSettings
): Promise<LineChartContent<any, string>> => {
    const [unit, value] = getStepInterval(input.step)

    return client
        .query<GetCaptureGroupInsightPreviewResult, GetCaptureGroupInsightPreviewVariables>({
            query: GET_CAPTURE_GROUP_INSIGHT_PREVIEW_GQL,
            variables: {
                input: {
                    query: input.query,
                    label: '',
                    repositoryScope: { repositories: input.repositories },
                    generatedFromCaptureGroups: true,
                    timeScope: { stepInterval: { unit, value } },
                },
            },
        })
        .then(({ data, error }) => {
            console.log(data)

            return LINE_CHART_WITH_HUGE_NUMBER_OF_LINES
        })
}
