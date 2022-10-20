import { formatISO } from 'date-fns'
import { escapeRegExp, groupBy } from 'lodash'

import { buildSearchURLQuery } from '@sourcegraph/shared/src/util/url'

import {
    GetInsightDataResult,
    InsightDataNode,
    InsightDataSeries,
    SearchPatternType,
} from '../../../../../../graphql-operations'
import { PageRoutes } from '../../../../../../routes.constants'
import { DATA_SERIES_COLORS_LIST } from '../../../../constants'
import {
    InsightType,
    isComputeInsight,
    BackendInsight,
    ComputeInsight,
    InsightFilters,
    SearchBasedInsightSeries,
} from '../../../../core'

import {
    InsightContentType,
    BackendInsightDTO,
    CategoricalChartContent,
    BackendInsightSeries,
} from './types'

const ALL_REPOS_POLL_INTERVAL = 30000
const SOME_REPOS_POLL_INTERVAL = 2000

export function insightPollingInterval(insight: BackendInsight): number {
    return insight.repositories.length > 0 ? SOME_REPOS_POLL_INTERVAL : ALL_REPOS_POLL_INTERVAL
}

export function isEntireInsightErrored (dto: BackendInsightDTO): boolean {
    if (dto.data.type === InsightContentType.Series) {
        return dto.data.series.every(series => series.status.errors.length > 0)
    }

    // Since in categorical insights there is no term such as series
    // if we've got even one error this means that the whole insight is
    // errored
    if (dto.data.type === InsightContentType.Categorical) {
        return dto.status.errors.length > 0
    }

    return false
}

/**
 * This selector prepares insight data in a way that it would be easy to work with later
 * in the insight consumers, our GQL API is far, far away from perfect, so we have to have
 * a client DTO models in order to transfer and use data objects without big knowledge about
 * API peculiarities
 */
export const parseBackendInsightResponse = (
    insight: BackendInsight,
    response?: GetInsightDataResult
): BackendInsightDTO | null => {
    if (!response) {
        return null
    }

    const { dataSeries, status } = getInsightData(response)
    const isInProgress = dataSeries.some(series => {
        const {
            status: { pendingJobs, backfillQueuedAt },
        } = series
        return pendingJobs > 0 || backfillQueuedAt === null
    })

    if (isComputeInsight(insight)) {
        return {
            // We have to tweak original logic around historical data since compute powered
            // insights have problem with generated data series status info
            // see https://github.com/sourcegraph/sourcegraph/issues/38893
            isInProgress: isInProgress || dataSeries.some(series => !series.label),
            status,
            data: {
                type: InsightContentType.Categorical,
                content: createComputeCategoricalChart(insight, dataSeries),
            },
        }
    }

    return {
        isInProgress,
        status,
        data: {
            type: InsightContentType.Series,
            series: createLineChartContent(insight, dataSeries),
        },
    }
}

function getInsightData(insightData: GetInsightDataResult): InsightDataNode {
    // It's safe to get a first element of insightViews because in case of GET_INSIGHT_DATA query
    // we specifically query insight by insight ID, it should be exactly one insight
    return insightData.insightViews.nodes[0]
}

/**
 * Generates line chart content for visx chart. Note that this function relies on the fact that
 * all series are indexed.
 */
function createLineChartContent(
    insight: BackendInsight,
    seriesData: InsightDataSeries[]
): BackendInsightSeries[] {
    const seriesDefinition = getParsedSeriesMetadata(insight, seriesData)
    const seriesDefinitionMap = Object.fromEntries<SearchBasedInsightSeries>(
        seriesDefinition.map(definition => [definition.id, definition])
    )

    return seriesData.map<BackendInsightSeries>(line => ({
        id: line.seriesId,
        status: line.status2,
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

function getParsedSeriesMetadata(insight: BackendInsight, seriesData: InsightDataSeries[]): SearchBasedInsightSeries[] {
    switch (insight.type) {
        case InsightType.SearchBased:
            return insight.series

        case InsightType.Compute: {
            return seriesData.map((generatedSeries, index) => ({
                id: generatedSeries.seriesId,
                name: generatedSeries.label,
                // TODO we don't know compute series contributions to each data items in dataset
                // see https://github.com/sourcegraph/sourcegraph/issues/38832
                query: '',
                stroke: DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length],
            }))
        }

        case InsightType.CaptureGroup: {
            const { query } = insight

            return seriesData.map((generatedSeries, index) => ({
                id: generatedSeries.seriesId,
                query,
                name: generatedSeries.label,
                stroke: DATA_SERIES_COLORS_LIST[index % DATA_SERIES_COLORS_LIST.length],
            }))
        }
    }
}

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

function createComputeCategoricalChart(insight: ComputeInsight, seriesData: InsightDataSeries[]): CategoricalChartContent {
    const seriesGroups = groupBy(
        seriesData.filter(series => series.label && series.points.length),
        series => series.label
    )

    // Group series result by seres name and sum up series value with the same name
    const groups = Object.keys(seriesGroups).map(key =>
        seriesGroups[key].reduce(
            (memo, series) => {
                memo.value += series.points.reduce((sum, datum) => sum + datum.value, 0)

                return memo
            },
            {
                name: seriesGroups[key][0].label,
                fill: insight.series[0]?.stroke ?? 'gray',
                value: 0,
            }
        )
    )

    return {
        data: groups,
        getDatumValue: datum => datum.value,
        getDatumColor: datum => datum.fill,
        getDatumName: datum => datum.name,
    }
}
