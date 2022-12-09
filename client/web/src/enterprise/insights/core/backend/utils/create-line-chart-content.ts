import { formatISO } from 'date-fns'
import { escapeRegExp } from 'lodash'

import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import { InsightDataSeries, SearchPatternType } from '../../../../../graphql-operations'
import { PageRoutes } from '../../../../../routes.constants'
import { BackendInsight, InsightFilters, SearchBasedInsightSeries } from '../../types'
import { BackendInsightDatum, BackendInsightSeries } from '../code-insights-backend-types'

import { getParsedSeriesMetadata } from './parse-series-metadata'

type SeriesDefinition = Record<string, SearchBasedInsightSeries>

interface LineChartContentInput {
    insight: BackendInsight
    seriesData: InsightDataSeries[]
    showError: boolean
}

/**
 * Generates line chart content for visx chart. Note that this function relies on the fact that
 * all series are indexed.
 */
export function createLineChartContent(input: LineChartContentInput): BackendInsightSeries<BackendInsightDatum>[] {
    const { insight, seriesData, showError } = input
    const seriesDefinition = getParsedSeriesMetadata(insight, seriesData)
    const seriesDefinitionMap: SeriesDefinition = Object.fromEntries<SearchBasedInsightSeries>(
        seriesDefinition.map(definition => [definition.id, definition])
    )

    return seriesData.map<BackendInsightSeries<BackendInsightDatum>>(line => ({
        // Encode all special symbols that may be in the original series such as
        // " symbol that breaks element query selectors in the chart where we use this id
        // for points and lines.
        // See https://github.com/sourcegraph/sourcegraph/issues/45376
        id: encodeURIComponent(line.seriesId),
        alerts: showError ? line.status.incompleteDatapoints : [],
        data: line.points.map((point, index) => ({
            dateTime: new Date(point.dateTime),
            value: point.value,
            link: generateLinkURL({
                point,
                previousPoint: line.points[index - 1],
                query: seriesDefinitionMap[line.seriesId].query,
                filters: insight.filters,
                repositories: insight.repositories,
            }),
        })),
        name: seriesDefinitionMap[line.seriesId]?.name ?? line.label,
        color: seriesDefinitionMap[line.seriesId]?.stroke,
        getYValue: datum => datum.value,
        getXValue: datum => datum.dateTime,
        getLinkURL: datum => datum.link,
    }))
}

/**
 * Minimal input type model for {@link createLineChartContent} function
 */
export type InsightDataSeriesData = Pick<InsightDataSeries, 'seriesId' | 'label' | 'points'>

interface GenerateLinkInput {
    query: string
    previousPoint?: { dateTime: string }
    point: { dateTime: string }
    repositories: string[]
    filters?: InsightFilters
}

export function generateLinkURL(input: GenerateLinkInput): string {
    const { query, point, previousPoint, filters, repositories } = input
    const { includeRepoRegexp = '', excludeRepoRegexp = '', context } = filters ?? {}

    const date = Date.parse(point.dateTime)

    // Use formatISO instead of toISOString(), because toISOString() always outputs UTC.
    // They mark the same point in time, but using the user's timezone makes the date string
    // easier to read (else the date component may be off by one day)
    const after = previousPoint ? formatISO(Date.parse(previousPoint.dateTime)) : ''
    const before = formatISO(date)

    const includeRepoFilter = includeRepoRegexp ? `repo:${includeRepoRegexp}` : ''
    const excludeRepoFilter = excludeRepoRegexp ? `-repo:${excludeRepoRegexp}` : ''

    const scopeRepoFilters = repositories.length > 0 ? `repo:^(${repositories.map(escapeRegExp).join('|')})$` : ''
    const contextFilter = context ? `context:${context}` : ''
    const repoFilter = `${includeRepoFilter} ${excludeRepoFilter}`
    const afterFilter = after ? `after:${after}` : ''
    const beforeFilter = `before:${before}`
    const dateFilters = `${afterFilter} ${beforeFilter}`
    const diffQuery = `${contextFilter} ${scopeRepoFilters} ${repoFilter} type:diff ${dateFilters} ${query}`
    const searchQueryParameter = buildSearchURLQuery(diffQuery, SearchPatternType.literal, false)

    return `${window.location.origin}${PageRoutes.Search}?${searchQueryParameter}`
}
