import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter, Keyword, Pattern } from '@sourcegraph/shared/src/search/query/token'

export interface Checks {
    isValidOperator: true | false | undefined
    isValidPatternType: true | false | undefined
    isNotRepo: true | false | undefined
    isNotCommitOrDiff: true | false | undefined
    isNoNewLines: true | false | undefined
}

export const searchQueryValidator = (value: string, touched: boolean): Checks => {
    if (!touched || !value || value.length === 0) {
        return {
            isValidOperator: undefined,
            isValidPatternType: undefined,
            isNotRepo: undefined,
            isNotCommitOrDiff: undefined,
            isNoNewLines: undefined,
        }
    }

    const tokens = scanSearchQuery(value)

    if (tokens.type === 'success') {
        const filters = tokens.term.filter(token => token.type === 'filter') as Filter[]
        const keywords = tokens.term.filter(token => token.type === 'keyword') as Keyword[]
        const patterns = tokens.term.filter(token => token.type === 'pattern') as Pattern[]

        const hasAnd = keywords.some(filter => filter.kind === 'and')
        const hasOr = keywords.some(filter => filter.kind === 'or')
        const hasNot = keywords.some(filter => filter.kind === 'not')

        const hasLiteralPattern = filters.some(
            filter =>
                resolveFilter(filter.field.value)?.type === FilterType.patterntype && filter.value?.value === 'literal'
        )

        const hasStructuralPattern = filters.some(
            filter =>
                resolveFilter(filter.field.value)?.type === FilterType.patterntype &&
                filter.value?.value === 'structural'
        )

        const hasRepo = filters.some(
            filter => resolveFilter(filter.field.value)?.type === FilterType.repo && filter.value
        )

        const hasCommit = filters.some(
            filter => resolveFilter(filter.field.value)?.type === FilterType.type && filter.value?.value === 'commit'
        )

        const hasDiff = filters.some(
            filter => resolveFilter(filter.field.value)?.type === FilterType.type && filter.value?.value === 'diff'
        )

        const hasNewLines = patterns.some(term => term.value === '\\n')

        return {
            isValidOperator: !hasAnd && !hasOr && !hasNot,
            isValidPatternType: !hasLiteralPattern && !hasStructuralPattern,
            isNotRepo: !hasRepo,
            isNotCommitOrDiff: !hasCommit && !hasDiff,
            isNoNewLines: !hasNewLines,
        }
    }

    return {
        isValidOperator: false,
        isValidPatternType: false,
        isNotRepo: false,
        isNotCommitOrDiff: false,
        isNoNewLines: false,
    }
}
