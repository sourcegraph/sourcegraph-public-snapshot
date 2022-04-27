import { ApolloClient, gql } from '@apollo/client'

import {
    GetCaptureGroupInsightPreviewResult,
    GetCaptureGroupInsightPreviewVariables,
} from '../../../../../../graphql-operations'
import { CaptureInsightSettings, SeriesChartContent } from '../../code-insights-backend-types'
import { generateLinkURL, InsightDataSeriesData } from '../../utils/create-line-chart-content'
import { getStepInterval } from '../utils/get-step-interval'

import { DATA_SERIES_COLORS_LIST, MAX_NUMBER_OF_SERIES } from './get-backend-insight-data/deserializators'

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
export interface CaptureGroupInsightDatum {
    dateTime: Date
    value: number | null
    link?: string
}

export const getCaptureGroupInsightsPreview = (
    client: ApolloClient<unknown>,
    input: CaptureInsightSettings
): Promise<SeriesChartContent<CaptureGroupInsightDatum>> => {
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

            // TODO Revisit live preview and dashboard insight resolver methods in order to
            // improve series data handling and manipulation
            const seriesMetadata = indexedSeries.map((generatedSeries, index) => ({
                id: generatedSeries.seriesId,
                name: generatedSeries.label,
                query: input.query,
                stroke: DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length],
            }))

            const seriesDefinitionMap = Object.fromEntries(
                seriesMetadata.map(definition => [definition.id, definition])
            )

            return {
                series: indexedSeries.map((line, index) => ({
                    id: line.seriesId,
                    data: line.points.map(point => ({
                        value: point.value,
                        dateTime: new Date(point.dateTime),
                        link: generateLinkURL({
                            previousPoint: line.points[index - 1],
                            series: seriesDefinitionMap[line.seriesId],
                            point,
                        }),
                    })),
                    name: line.label,
                    color: DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length],
                    getLinkURL: datum => datum.link,
                    getYValue: datum => datum.value,
                    getXValue: datum => datum.dateTime,
                })),
            }
        })
}
