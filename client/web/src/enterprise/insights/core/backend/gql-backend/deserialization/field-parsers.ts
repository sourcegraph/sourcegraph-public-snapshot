import type { Duration } from 'date-fns'

import {
    type InsightViewNode,
    type TimeIntervalStepInput,
    SeriesSortDirection,
    SeriesSortMode,
    TimeIntervalStepUnit,
} from '../../../../../../graphql-operations'
import { MAX_NUMBER_OF_SAMPLES } from '../../../../constants'
import type { InsightFilters, InsightSeriesDisplayOptions } from '../../../types/insight/common'

export function getDurationFromStep(step: TimeIntervalStepInput): Duration {
    switch (step.unit) {
        case TimeIntervalStepUnit.HOUR: {
            return { hours: step.value }
        }
        case TimeIntervalStepUnit.DAY: {
            return { days: step.value }
        }
        case TimeIntervalStepUnit.WEEK: {
            return { weeks: step.value }
        }
        case TimeIntervalStepUnit.MONTH: {
            return { months: step.value }
        }
        case TimeIntervalStepUnit.YEAR: {
            return { years: step.value }
        }
    }
}

type ResponseSeriesDisplayOptions = InsightViewNode['defaultSeriesDisplayOptions']
type ResponseAppliedFilters = InsightViewNode['defaultFilters']

export function getParsedFilters(
    rawAppliedFilters: ResponseAppliedFilters,
    rawDisplayOptions: ResponseSeriesDisplayOptions
): InsightFilters {
    return {
        includeRepoRegexp: rawAppliedFilters.includeRepoRegex ?? '',
        excludeRepoRegexp: rawAppliedFilters.excludeRepoRegex ?? '',
        context: rawAppliedFilters.searchContexts?.[0] ?? '',
        seriesDisplayOptions: getParsedSeriesOption(rawDisplayOptions),
    }
}

function getParsedSeriesOption(response: ResponseSeriesDisplayOptions): InsightSeriesDisplayOptions {
    const { limit, numSamples, sortOptions } = response

    // Have to check zero value because of backend problem (it always returns 0 when
    // numSamples isn't applied
    const parsedNumSamples =
        numSamples !== null && numSamples !== 0 ? Math.min(numSamples, MAX_NUMBER_OF_SAMPLES) : null
    const parsedSortOptions = {
        mode: sortOptions.mode ?? SeriesSortMode.RESULT_COUNT,
        direction: sortOptions.direction ?? SeriesSortDirection.DESC,
    }

    return {
        limit,
        numSamples: parsedNumSamples,
        sortOptions: parsedSortOptions,
    }
}

type ResponseRepositoryDefinition = InsightViewNode['repositoryDefinition']

interface InsightRepositories {
    repoSearch: string
    repositories: string[]
}

export function getInsightRepositories(response: ResponseRepositoryDefinition): InsightRepositories {
    if (response.__typename === 'InsightRepositoryScope') {
        return {
            repoSearch: '',
            repositories: response.repositories,
        }
    }

    return {
        repoSearch: response.search,
        repositories: [],
    }
}
