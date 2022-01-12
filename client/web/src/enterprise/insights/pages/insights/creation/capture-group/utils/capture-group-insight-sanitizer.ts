import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter } from '@sourcegraph/shared/src/search/query/token'

import { getSanitizedRepositories } from '../../../../../components/creation-ui-kit/sanitizers/repositories'
import { CaptureGroupInsight, InsightExecutionType, InsightType } from '../../../../../core/types'
import { CaptureGroupFormFields } from '../types'

export function getSanitizedCaptureGroupInsight(values: CaptureGroupFormFields): CaptureGroupInsight {
    return {
        title: values.title.trim(),
        query: getSanitizedCaptureQuery(values.groupSearchQuery.trim()),
        repositories: getSanitizedRepositories(values.repositories),
        viewType: InsightType.CaptureGroup,
        type: InsightExecutionType.Backend,
        id: '',
        visibility: '',
        step: { [values.step]: +values.stepValue },
        filters: { excludeRepoRegexp: '', includeRepoRegexp: '' },
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
