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
        series: series
            .map<Series<BackendInsightDatum>>(line => ({
                id: line.seriesId,
                data: line.points.map((point, index) => ({
                    dateTime: new Date(point.dateTime),
                    value: point.value,
                    link: generateLinkURL({
                        previousPoint: line.points[index - 1],
                        series: seriesDefinitionMap[line.seriesId],
                        point,
                        includeRepoRegexp,
                        excludeRepoRegexp,
                    }),
                })),
                name: seriesDefinitionMap[line.seriesId]?.name ?? line.label,
                color: seriesDefinitionMap[line.seriesId]?.stroke,
                getYValue: datum => datum.value,
                getXValue: datum => datum.dateTime,
                getLinkURL: datum => datum.link,
            }))
            .sort((a, b) => semanticSort(a.name, b.name)),
    }
}

interface GenerateLinkInput {
    series: SearchBasedInsightSeries
    previousPoint?: { dateTime: string }
    point: { dateTime: string }
    includeRepoRegexp?: string
    excludeRepoRegexp?: string
}

export function generateLinkURL(input: GenerateLinkInput): string {
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
