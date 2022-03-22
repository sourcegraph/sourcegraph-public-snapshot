import { ApolloClient, gql } from '@apollo/client'
import { startCase } from 'lodash'
import openColor from 'open-color'
import { LineChartContent } from 'sourcegraph'

import {
    GetCaptureGroupInsightPreviewResult,
    GetCaptureGroupInsightPreviewVariables,
} from '../../../../../../graphql-operations'
import { CaptureInsightSettings } from '../../code-insights-backend-types'
import { getDataPoints, InsightDataSeriesData } from '../../utils/create-line-chart-content'
import { getStepInterval } from '../utils/get-step-interval'

import { MAX_NUMBER_OF_SERIES } from './get-backend-insight-data/deserializators'

const SERIES_COLORS = Object.keys(openColor)
    .filter(name => name !== 'white' && name !== 'black' && name !== 'gray')
    .map(name => ({ name: startCase(name), color: `var(--oc-${name}-7)` }))

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
                    timeScope: { stepInterval: { unit, value: +value } },
                },
            },
        })
        .then(({ data, error }) => {
            if (error) {
                throw error
            }

            const { searchInsightLivePreview: series } = data

            if (series.length === 0) {
                throw new Error('Found no matches')
            }

            // Extend series with synthetic index based series id
            const indexedSeries = series.slice(0, MAX_NUMBER_OF_SERIES).map<InsightDataSeriesData>((series, index) => ({
                seriesId: `${index}`,
                ...series,
            }))

            return {
                chart: 'line',
                data: getDataPoints(indexedSeries),
                series: indexedSeries.map((series, index) => ({
                    dataKey: series.seriesId,
                    name: series.label,
                    stroke: SERIES_COLORS[index % SERIES_COLORS.length].color,
                })),
                xAxis: {
                    dataKey: 'dateTime',
                    scale: 'time',
                    type: 'number',
                },
            }
        })
}
