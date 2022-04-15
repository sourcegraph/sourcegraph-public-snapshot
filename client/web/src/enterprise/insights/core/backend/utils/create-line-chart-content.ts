import { formatISO } from 'date-fns'

import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { Series } from '../../../../../charts'
import { InsightDataSeries, SearchPatternType } from '../../../../../graphql-operations'
import { semanticSort } from '../../../../../insights/utils/semantic-sort'
import { PageRoutes } from '../../../../../routes.constants'
import { InsightFilters, SearchBasedInsightSeries } from '../../types'
import { BackendInsightDatum, SeriesChartContent } from '../code-insights-backend-types'

type SeriesDefinition = Record<string, SearchBasedInsightSeries>

/**
 * Minimal input type model for {@link createLineChartContent} function
 */
export type InsightDataSeriesData = Pick<InsightDataSeries, 'seriesId' | 'label' | 'points'>

/**
 * Generates line chart content for visx chart. Note that this function relies on the fact that
 * all series are indexed.
 *
 * @param series - insight series with points data
 * @param seriesDefinition - insight definition with line settings (color, name, query)
 * @param filters - insight drill-down filters
 */
export function createLineChartContent(
    series: InsightDataSeriesData[],
    seriesDefinition: SearchBasedInsightSeries[] = [],
    filters?: InsightFilters
): SeriesChartContent<BackendInsightDatum> {
    const seriesDefinitionMap: SeriesDefinition = Object.fromEntries<SearchBasedInsightSeries>(
        seriesDefinition.map(definition => [definition.id, definition])
    )

    const { includeRepoRegexp = '', excludeRepoRegexp = '' } = filters ?? {}

    return {
        data: getDataPoints({ series, seriesDefinitionMap, excludeRepoRegexp, includeRepoRegexp }),
        series: series
            .map<Series<BackendInsightDatum>>(line => ({
                dataKey: line.seriesId,
                name: seriesDefinitionMap[line.seriesId]?.name ?? line.label,
                color: seriesDefinitionMap[line.seriesId]?.stroke,
                getLinkURL: datum => `${datum[getLinkKey(line.seriesId)]}`,
            }))
            .sort((a, b) => semanticSort(a.name, b.name)),
        getXValue: datum => new Date(datum.dateTime),
    }
}

interface GetDataPointsInput {
    series: InsightDataSeriesData[]
    seriesDefinitionMap: SeriesDefinition
    includeRepoRegexp: string
    excludeRepoRegexp: string
}

/**
 * Groups data series by dateTime (x-axis) of each series
 */
export function getDataPoints(input: GetDataPointsInput): BackendInsightDatum[] {
    const { series, seriesDefinitionMap, includeRepoRegexp, excludeRepoRegexp } = input
    const dataByXValue = new Map<string, BackendInsightDatum>()

    for (const line of series) {
        for (const [index, point] of line.points.entries()) {
            let dataObject = dataByXValue.get(point.dateTime)
            if (!dataObject) {
                dataObject = {
                    dateTime: Date.parse(point.dateTime),
                    // Initialize all series to null (empty chart) value
                    ...Object.fromEntries(series.map(line => [line.seriesId, null])),
                }
                dataByXValue.set(point.dateTime, dataObject)
            }
            dataObject[line.seriesId] = point.value
            dataObject[getLinkKey(line.seriesId)] = generateLinkURL({
                previousPoint: line.points[index - 1],
                series: seriesDefinitionMap[line.seriesId],
                point,
                includeRepoRegexp,
                excludeRepoRegexp,
            })
        }
    }

    return [...dataByXValue.values()]
}

interface GenerateLinkInput {
    series: SearchBasedInsightSeries
    previousPoint?: { dateTime: string }
    point: { dateTime: string }
    includeRepoRegexp: string
    excludeRepoRegexp: string
}

function generateLinkURL(input: GenerateLinkInput): string {
    const { series, point, previousPoint, includeRepoRegexp, excludeRepoRegexp } = input

    const date = Date.parse(point.dateTime)

    // Use formatISO instead of toISOString(), because toISOString() always outputs UTC.
    // They mark the same point in time, but using the user's timezone makes the date string
    // easier to read (else the date component may be off by one day)
    const after = previousPoint ? formatISO(Date.parse(previousPoint.dateTime)) : ''
    const before = formatISO(date)

    const includeRepoFilter = includeRepoRegexp ? `repo:${includeRepoRegexp}` : ''
    const excludeRepoFilter = excludeRepoRegexp ? `-repo:${excludeRepoRegexp}` : ''

    const repoFilter = `${includeRepoFilter} ${excludeRepoFilter}`
    const afterFilter = after ? `after:${after}` : ''
    const beforeFilter = `before:${before}`
    const dateFilters = `${afterFilter} ${beforeFilter}`
    const diffQuery = `${repoFilter} type:diff ${dateFilters} ${series.query}`
    const searchQueryParameter = buildSearchURLQuery(diffQuery, SearchPatternType.literal, false)

    return `${window.location.origin}${PageRoutes.Search}?${searchQueryParameter}`
}

export function getLinkKey(seriesId: string): string {
    return `${seriesId}:link`
}
