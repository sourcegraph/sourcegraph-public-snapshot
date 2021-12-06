import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter, Keyword } from '@sourcegraph/shared/src/search/query/token'

const regexCheck = (value: string): boolean => {
    try {
        new RegExp(value)
        return true
    } catch {
        return false
    }
}

export interface Checks {
    isValidRegex: boolean
    isValidOperator: boolean
    isValidPatternType: boolean
    isNotRepoOrFile: boolean
    isNotCommitOrDiff: boolean
    isNoRepoFilter: boolean
}

export const searchQueryValidator = (value: string, checks: Checks): Checks => {
    const tokens = scanSearchQuery(value)
    const validatedChecks = { ...checks }

    if (tokens.type === 'success') {
        const filters = tokens.term.filter(token => token.type === 'filter') as Filter[]
        const keywords = tokens.term.filter(token => token.type === 'keyword') as Keyword[]

        const hasRepo = filters.some(
            filter => resolveFilter(filter.field.value)?.type === FilterType.repo && filter.value
        )

        const hasFile = filters.some(
            filter => resolveFilter(filter.field.value)?.type === FilterType.file && filter.value
        )

        const hasLiteralPattern = filters.some(
            filter =>
                resolveFilter(filter.field.value)?.type === FilterType.patterntype && filter.value?.value === 'literal'
        )

        const hasStructuralPattern = filters.some(
            filter =>
                resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                filter.value?.value === 'structural'
        )

        const hasCommit = filters.some(
            filter => resolveFilter(filter.field.value)?.type === FilterType.type && filter.value?.value === 'commit'
        )

        const hasDiff = filters.some(
            filter => resolveFilter(filter.field.value)?.type === FilterType.type && filter.value?.value === 'diff'
        )

        const hasAnd = keywords.some(filter => filter.kind === 'and')
        const hasOr = keywords.some(filter => filter.kind === 'or')

        validatedChecks.isValidRegex = regexCheck(value)
        validatedChecks.isNotCommitOrDiff = !hasCommit && !hasDiff
        validatedChecks.isNoRepoFilter = !hasRepo
        validatedChecks.isNotRepoOrFile = !hasRepo && !hasFile
        validatedChecks.isValidPatternType = !hasLiteralPattern && !hasStructuralPattern
        validatedChecks.isValidOperator = !hasAnd && !hasOr
    }

    return validatedChecks
}
