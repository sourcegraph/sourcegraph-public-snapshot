import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter } from '@sourcegraph/shared/src/search/query/token'

import { SeriesSortDirection, SeriesSortMode } from '../../../../../../../graphql-operations'
import { getSanitizedRepositories } from '../../../../../components'
import { MAX_NUMBER_OF_SERIES } from '../../../../../constants'
import { InsightExecutionType, InsightType, MinimalCaptureGroupInsightData } from '../../../../../core'
import { CaptureGroupFormFields } from '../types'

export function getSanitizedCaptureGroupInsight(values: CaptureGroupFormFields): MinimalCaptureGroupInsightData {
    return {
        title: values.title.trim(),
        query: getSanitizedCaptureQuery(values.groupSearchQuery.trim()),
        type: InsightType.CaptureGroup,
        executionType: InsightExecutionType.Backend,
        step: { [values.step]: +values.stepValue },
        repositories: values.allRepos ? [] : getSanitizedRepositories(values.repositories),
        filters: {
            includeRepoRegexp: '',
            excludeRepoRegexp: '',
            context: '',
            seriesDisplayOptions: {
                limit: `${MAX_NUMBER_OF_SERIES}`,
                sortOptions: {
                    direction: SeriesSortDirection.DESC,
                    mode: SeriesSortMode.RESULT_COUNT,
                },
            },
        },
        dashboards: [],
        seriesDisplayOptions: {},
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
