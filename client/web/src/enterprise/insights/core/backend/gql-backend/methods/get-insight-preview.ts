import { ApolloClient, gql } from '@apollo/client'

import { GetInsightPreviewResult, GetInsightPreviewVariables } from '../../../../../../graphql-operations'
import { BackendInsightDatum, InsightPreviewSettings, SeriesChartContent } from '../../code-insights-backend-types'
import { generateLinkURL, InsightDataSeriesData } from '../../utils/create-line-chart-content'
import { getStepInterval } from '../utils/get-step-interval'

import { DATA_SERIES_COLORS_LIST, MAX_NUMBER_OF_SERIES } from './get-backend-insight-data/deserializators'

const GET_INSIGHT_PREVIEW_GQL = gql`
    query GetInsightPreview($input: SearchInsightPreviewInput!) {
        searchInsightPreview(input: $input) {
            points {
                dateTime
                value
            }
            label
        }
    }
`

export const getInsightsPreview = (
    client: ApolloClient<unknown>,
    input: InsightPreviewSettings
): Promise<SeriesChartContent<BackendInsightDatum>> => {
    const [unit, value] = getStepInterval(input.step)

    // inputMetadata creates a lookup so that the correct color can be later applied to the preview series
    const inputMetadata = Object.fromEntries(
        input.series.map((previewSeries, index) => [`${previewSeries.label}-${index}`, previewSeries])
    )

    // TODO(insights): inputMetadata and this function need to be re-evaluated in the future if/when support for
    // mixing series types in a single insight is possible
    function getColorForSeries(label: string, index: number): string {
        return (
            inputMetadata[`${label}-${index}`]?.stroke ||
            DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length]
        )
    }

    return client
        .query<GetInsightPreviewResult, GetInsightPreviewVariables>({
            query: GET_INSIGHT_PREVIEW_GQL,
            variables: {
                input: {
                    repositoryScope: { repositories: input.repositories },
                    timeScope: { stepInterval: { unit, value: +value } },
                    series: input.series.map(previewSeries => ({
                        generatedFromCaptureGroups: previewSeries.generatedFromCaptureGroup,
                        query: previewSeries.query,
                        label: previewSeries.label,
                    })),
                },
            },
        })
        .then(({ data, error }) => {
            if (error) {
                throw error
            }

            const { searchInsightPreview: series } = data

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
                query: inputMetadata[generatedSeries.label]?.query || '',
                stroke: getColorForSeries(generatedSeries.label, index),
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
                    color: getColorForSeries(line.label, index),
                    getLinkURL: datum => datum.link,
                    getYValue: datum => datum.value,
                    getXValue: datum => datum.dateTime,
                })),
            }
        })
}
