import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter } from '@sourcegraph/shared/src/search/query/token'

import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../graphql-operations'
import { InsightType, type MinimalCaptureGroupInsightData } from '../../../../../core'
import type { CaptureGroupFormFields } from '../types'

export function getSanitizedCaptureGroupInsight(values: CaptureGroupFormFields): MinimalCaptureGroupInsightData {
    return {
        title: values.title.trim(),
        query: getSanitizedCaptureQuery(values.groupSearchQuery.trim()),
        type: InsightType.CaptureGroup,
        step: { [values.step]: +values.stepValue },
        repoQuery: values.repoMode === 'search-query' ? values.repoQuery.query : '',
        repositories: values.repoMode === 'urls-list' ? values.repositories : [],

        filters: {
            includeRepoRegexp: '',
            excludeRepoRegexp: '',
            context: '',
            seriesDisplayOptions: {
                limit: null,
                numSamples: null,
                sortOptions: {
                    direction: SeriesSortDirection.DESC,
                    mode: SeriesSortMode.RESULT_COUNT,
                },
            },
        },
        dashboards: [],
    }
}

export const getSanitizedCaptureQuery = (query: string): string => {
    const tokens = scanSearchQuery(query)

    if (tokens.type === 'success') {
        const filters = tokens.term.filter(token => token.type === 'filter') as Filter[]

        const hasRegExpPattern = filters.some(
            filter =>
                resolveFilter(filter.field.value)?.type === FilterType.patterntype && filter.value?.value === 'regexp'
        )

        return hasRegExpPattern ? query : `patterntype:regexp ${query}`
    }

    return query
}
